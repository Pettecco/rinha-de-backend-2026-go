# Rinha de Backend 2026 — Go

Minha solução pro desafio de detecção de fraude com busca vetorial. O objetivo é simples: receber transações via HTTP, converter em vetores 14-D, buscar as 5 mais parecidas num índice de 3 milhões de referências, e devolver um score de fraude.

Tudo rodando com **1 CPU e 350 MB de RAM**, dividido entre nginx + 2 instâncias de API.

## Como rodar

### Pré-requisitos

- Docker + Docker Compose
- k6 (opcional, pra rodar o benchmark local)

### Passo a passo

```bash
# 1. Clonar o repo
git clone https://github.com/SEU_USER/rinha-backend-2026-go.git
cd rinha-backend-2026-go

# 2. Build (gera o índice IVF a partir dos 3M de vetores)
#    Primeira vez demora ~3min. Depois o Docker cacheia.
make build

# 3. Subir tudo (nginx + 2x API)
make up

# 4. Verificar se está no ar
curl http://localhost:9999/ready

# 5. Rodar o benchmark (precisa do k6 instalado)
make bench

# 6. Ver logs
make logs

# 7. Derrubar
make down
```

### Sem Makefile?

```bash
docker compose build
docker compose up -d
k6 run test/test.js
docker compose down
```

## O que acontece por dentro

A API recebe JSON, transforma em vetor normalizado, busca num índice IVF pré-construído com mmap, e devolve a resposta em ~3-5ms.

### O fluxo

1. **Parse** — `jsonparser.EachKey` varre o JSON uma vez, extraindo os campos direto em `[]byte` (zero alocações de string).
2. **Vetorização** — 14 dimensões são calculadas: amount normalizado, hora do dia, dia da semana, distância, etc. Timestamps são parseados com aritmética pura (sem `time.Parse`).
3. **Quantização** — float64 vira int16 (escala 10000) pra caber menos na cache e processar mais rápido.
4. **Busca IVF** — O índice tem 4096 clusters. O estágio rápido varre 8, o completo varre 28 (só quando o resultado tá no limite entre aprovar/reprovar).
5. **Distância SIMD** — Em CPUs Intel/AMD, 8 dimensões são processadas de uma vez com AVX2. Em ARM, fallback escalar.
6. **TopK** — Array plano de 5 posições com scan linear. Sem heap, sem ponteiros.
7. **Resposta** — JSON pré-formatado, escrito direto no buffer.

### Por que é rápido

| Técnica            | Impacto                                    |
| ------------------ | ------------------------------------------ |
| fasthttp           | Zero alloc por request                     |
| UDS (sem TCP)      | ~30µs a menos por request                  |
| mmap do índice     | Sem cópia pra memória, o OS gerencia       |
| PreTouch           | Páginas carregadas antes do tráfego chegar |
| sync.Pool          | Buffer de 32KB reusado, sem GC pressure    |
| SIMD AVX2          | 8 dimensões por instrução                  |
| Loops desenrolados | Sem overhead de loop no hot path           |

## Estrutura do projeto

```
cmd/api/              API HTTP (fasthttp, UDS)
cmd/indexer/          Gera o ivf.bin a partir do dataset
internal/
  vector/             Vetorização do payload
  quantize/           Quantização float64 → int16
  ivf/                Busca IVF (runtime)
  ivf/builder/        Build do índice (k-means, writer)
  simd/               AVX2 intrinsics + fallback escalar
  consts/             Constantes globais
resources/            Dataset (references.json.gz, etc)
test/                 Script k6 + dados de teste
```

## Configuração

Variáveis de ambiente no `docker-compose.yml`:

| Variável       | Padrão | O que faz                         |
| -------------- | ------ | --------------------------------- |
| `N_PROBE_FAST` | 8      | Clusters no estágio rápido        |
| `N_PROBE_FULL` | 28     | Clusters no estágio completo      |
| `GOMAXPROCS`   | 1      | Threads do Go (1 evita contenção) |
| `GOMEMLIMIT`   | 150MiB | Limite soft de heap               |
