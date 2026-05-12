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
WORKDIR /app
COPY --from=builder /src/api .
COPY --from=builder /src/dataset/ivf.bin /data/ivf.bin
COPY --from=builder /src/dataset/normalization.json /data/
COPY --from=builder /src/dataset/mcc_risk.json /data/
ENV DATA_DIR=/data
EXPOSE 8080
CMD ["./api"]
