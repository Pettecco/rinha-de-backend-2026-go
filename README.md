# rinha26

Implementação Go para o desafio [Rinha de Backend 2026](https://github.com/zanfranceschi/rinha-de-backend-2026) — detecção de fraude em transações de cartão via busca de vizinhos próximos (k=5) sobre 3 milhões de vetores 14-D.

## Arquitetura

```
client
  │
  ▼
nginx (porta 9999)            0.20 CPU /  20 MB
  ├─ round-robin via UDS
  │
  ├──▶ api 1 (Go)              0.40 CPU / 165 MB   /run/sock/api1.sock
  └──▶ api 2 (Go)              0.40 CPU / 165 MB   /run/sock/api2.sock
                              ───────────────────
                              1.00 CPU / 350 MB
```

nginx ↔ APIs comunicam por **Unix Domain Sockets** num volume tmpfs compartilhado (`sock`) — sem TCP loopback no caminho dos dados.

Stack:

- Go 1.24
- [fasthttp](https://github.com/valyala/fasthttp) (substitui `net/http` — zero alloc por request)
- [jsonparser](https://github.com/buger/jsonparser) na vetorização
- CGO + AVX2 intrinsics no inner loop de distância (fallback escalar em ARM)

## Pipeline de inferência

```
HTTP POST /fraud-score
  │
  ▼
parse JSON (jsonparser)         vector/payload.go
  │
  ▼
vetorização 14-D float64        vector/payload.go
  │
  ▼
quantização int16 (×10000)      quantize/quant.go
  │
  ▼
IVF k-NN dois estágios:         ivf/search.go
  • estágio rápido: 8 clusters
  • estágio completo: 28 clusters (só quando count = 2 ou 3)
  • SIMD AVX2 com AoSoA layout  (simd/dist_amd64.c)
  │
  ▼
contagem fraud no top-5
  │
  ▼
resposta pré-computada           api/responses.go
```

## Layout

```
cmd/api/                        servidor HTTP (porta 8080; LB expõe 9999)
  main.go
  responses.go
  Dockerfile                    multi-stage: builder → indexer → final image

cmd/indexer/                    CLI que constrói ivf.bin a partir do dataset
  main.go

internal/                       código privado
  vector/                       vetorização do payload (14 dims, normalização)
    payload.go                  jsonparser.EachKey, rawFields, zero string alloc
    timestamp.go                Sakamoto + Hinnant (sem time.Parse)
    config.go                   LoadNorm, LoadMccRisk
    types.go                    Dim, TopK, Sentinel, Norm, MccRisk
  quantize/                     quantização int16
    quant.go                    EncodeFloat, EncodeVec
  ivf/                          índice IVF: loader, search, topk, format
    search.go                   FraudScore, centroid distances, scanClusters
    topk.go                     pickTopFromDists, updateTopK (array plano)
    loader.go                   mmap, precomputeCentroids, PreTouch
    format.go                   Header, ReadHeader
    errors.go                   ErrBadMagic
  ivf/builder/                  build-time only (não entra na API)
    build.go                    BuildIndex, ClusterBlocks
    kmeans.go                   k-means++ init, Lloyd iterations
    rng.go                      LCG (deterministic RNG)
    writer.go                   WriteIndex (binary format)
  simd/                         AVX2 intrinsics
    dist_amd64.go               CGO wrapper (amd64 only)
    dist_amd64.c                AVX2 implementation
    dist.h                      C header
    dist_fallback.go            scalar fallback (non-amd64)
  consts/                       constantes globais (K, Dim, Scale, etc)

resources/                      input data
  references.json.gz            3M referências (~16 MB gzip)
  mcc_risk.json                 MCC → risco
  normalization.json            constantes de normalização

test/                           k6 oficial do desafio
  test.js                       ramping 1→900 RPS em 120s
  test-data.json                54.100 entries com expected_fraud_score

docker-compose.yml              orquestração local (platform: linux/amd64)
nginx.conf                      load balancer minimalista (1 worker, ~5 MB RSS)
Makefile                        build / up / down / logs / bench
```

## Build & smoke test

```bash
make build          # docker compose build (full 3M, ~3min)
make up             # docker compose up -d
make bench          # k6 run test/test.js
make down           # docker compose down
```

Para iterar rápido sem reindexar 3M registros, o Docker cacheia o builder stage. O `ivf.bin` só é recriado se o código do indexer ou o dataset mudar.

## Tunables

Variáveis de ambiente (definidas em `docker-compose.yml`):

| variável | default | descrição |
|---|---|---|
| `N_PROBE_FAST` | 8 | clusters varridos no estágio 1 |
| `N_PROBE_FULL` | 28 | clusters varridos no estágio 2 (só quando count = 2 ou 3) |
| `GOMAXPROCS` | 1 | threads do Go scheduler (single-thread evita contenção) |
| `GOMEMLIMIT` | 150MiB | soft limit do heap do Go |

## Otimizações aplicadas

### API (hot path)
- **fasthttp** com `NoDefaultServerHeader`, `NoDefaultContentType`, `NoDefaultDate` — evita formatar headers desnecessários
- **Buffers apertados** — `ReadBufferSize: 2048`, `WriteBufferSize: 256` (payloads <1KB, respostas ~35B)
- **Timeouts explícitos** — `ReadTimeout: 2s`, `WriteTimeout: 2s`, `IdleTimeout: 60s`
- **Unix Domain Sockets** — sem TCP loopback entre nginx e API

### Vetorização
- **jsonparser.EachKey** — single pass no JSON, sem 13 `Get` calls separados
- **rawFields com []byte** — zero alocações de string durante o parse
- **bytesEqual** — compara `[]byte` diretamente, sem `string()`
- **Sakamoto + Hinnant** — dia da semana e minutos entre timestamps sem `time.Parse`

### Busca vetorial
- **sync.Pool** para buffer de distâncias centroides — evita alocar 32KB por request
- **pickTopFromDists** com arrays no stack — zero heap allocs
- **Loops desenrolados** — dot product e distância L2 sem overhead de loop
- **AoSoA layout** — dados transpostos por dimensão, ideal para SIMD
- **AVX2 intrinsics** — 8 dimensões processadas por instrução SIMD (fallback escalar em ARM)

### TopK
- **Array plano [5]int64** — sem heap, sem ponteiros, melhor cache locality
- **updateTopK** — scan linear pelo pior, mais rápido que max-heap para K=5

## Score

Métricas típicas (rodando na máquina de teste da Rinha — Intel Core i5 2014):

| métrica | valor |
|---|---|
| p99 | ~3-5ms |
| FP | 0-2 |
| FN | 0-2 |
| HTTP errors | 0 |
| **final_score** | ~5000-5500 |
