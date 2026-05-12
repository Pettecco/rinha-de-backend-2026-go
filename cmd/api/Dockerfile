FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN mkdir -p dataset && cp resources/references.json.gz dataset/
RUN go build -o indexer ./cmd/indexer
RUN go build -o api ./cmd/api
RUN ./indexer

FROM alpine:3.20
WORKDIR /app/dataset
COPY --from=builder /src/api /app/
COPY --from=builder /src/dataset/ivf.bin .
COPY --from=builder /src/dataset/normalization.json .
COPY --from=builder /src/dataset/mcc_risk.json .
WORKDIR /app
EXPOSE 8080
CMD ["./api"]
