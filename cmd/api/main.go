package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"rinha26/internal/ivf"
	"rinha26/internal/vec"
)

var responses = [6][]byte{
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 40\r\nConnection: keep-alive\r\n\r\n{\"approved\":true,\"fraud_score\":0.0}"),
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 40\r\nConnection: keep-alive\r\n\r\n{\"approved\":true,\"fraud_score\":0.2}"),
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 40\r\nConnection: keep-alive\r\n\r\n{\"approved\":true,\"fraud_score\":0.4}"),
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 41\r\nConnection: keep-alive\r\n\r\n{\"approved\":false,\"fraud_score\":0.6}"),
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 41\r\nConnection: keep-alive\r\n\r\n{\"approved\":false,\"fraud_score\":0.8}"),
	[]byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 40\r\nConnection: keep-alive\r\n\r\n{\"approved\":false,\"fraud_score\":1.0}"),
}

var fallback = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 41\r\nConnection: keep-alive\r\n\r\n{\"approved\":false,\"fraud_score\":0.6}")

var index *ivf.Index
var norm *vec.Norm
var mccRisk map[string]float64
var nProbeFast int
var nProbeFull int

func handler(ctx *fasthttp.RequestCtx) {
	if ctx.IsGet() {
		if string(ctx.Path()) == "/ready" {
			ctx.SetStatusCode(200)
			return
		}
		ctx.SetStatusCode(404)
		return
	}

	if !ctx.IsPost() || string(ctx.Path()) != "/fraud-score" {
		ctx.SetStatusCode(404)
		return
	}

	body := ctx.PostBody()
	if len(body) == 0 {
		ctx.Write(fallback)
		return
	}

	var query [14]float64
	if err := vec.FromPayload(body, &query, norm, mccRisk); err != nil {
		ctx.Write(fallback)
		return
	}

	fraudCount := index.FraudScore(query, nProbeFast, nProbeFull)
	ctx.Write(responses[fraudCount])
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func main() {
	var err error

	norm, err = vec.LoadNorm("dataset/normalization.json")
	if err != nil {
		log.Fatalf("load norm: %v", err)
	}

	mccRisk, err = vec.LoadMccRisk("dataset/mcc_risk.json")
	if err != nil {
		log.Fatalf("load mcc: %v", err)
	}

	idxPath := os.Getenv("INDEX_PATH")
	if idxPath == "" {
		idxPath = "dataset/ivf.bin"
	}

	index, err = ivf.Open(idxPath)
	if err != nil {
		log.Fatalf("open index: %v", err)
	}
	defer index.Close()

	index.PreTouch()

	nProbeFast = envInt("N_PROBE_FAST", 8)
	nProbeFull = envInt("N_PROBE_FULL", 28)

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("listening on %s (nprobe fast=%d, full=%d)", addr, nProbeFast, nProbeFull)

	server := &fasthttp.Server{
		Handler:                       handler,
		Concurrency:                   256,
		DisableHeaderNamesNormalizing: true,
		ReadBufferSize:                4096,
		WriteBufferSize:               1024,
	}

	if strings.HasPrefix(addr, "/") {
		os.Remove(addr)
		ln, err := net.Listen("unix", addr)
		if err != nil {
			log.Fatalf("listen unix: %v", err)
		}
		os.Chmod(addr, 0666)
		log.Fatal(server.Serve(ln))
	} else {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("listen tcp: %v", err)
		}
		log.Fatal(server.Serve(ln))
	}
}
