package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/valyala/fasthttp"
	"rinha26/internal/ivf"
	"rinha26/internal/vector"
)

var index *ivf.Index
var norm *vector.Norm
var mccRisk vector.MccRisk
var isReady bool

const (
	nProbeFast = 8
	nProbeFull = 28
)

func handler(ctx *fasthttp.RequestCtx) {
	if ctx.IsGet() {
		if isReady {
			ctx.SetStatusCode(200)
		} else {
			ctx.SetStatusCode(503)
		}
		return
	}

	var query [14]float64
	if err := vector.FromPayload(ctx.PostBody(), norm, mccRisk, &query); err != nil {
		ctx.SetStatusCode(400)
		return
	}

	fraudCount := index.FraudScore(query, nProbeFast, nProbeFull)
	ctx.SetContentType("application/json")
	ctx.Write(responses[fraudCount])
}

func main() {
	var err error

	norm, err = vector.LoadNorm("dataset/normalization.json")
	if err != nil {
		log.Fatalf("load norm: %v", err)
	}

	mccRisk, err = vector.LoadMccRisk("dataset/mcc_risk.json")
	if err != nil {
		log.Fatalf("load mcc: %v", err)
	}

	index, err = ivf.Open("dataset/ivf.bin")
	if err != nil {
		log.Fatalf("open index: %v", err)
	}
	defer index.Close()

	index.PreTouch()

	isReady = true

	addr := os.Getenv("LISTEN_ADDR")

	log.Printf("listening on %s", addr)

	server := &fasthttp.Server{
		Handler:                  handler,
		Name:                     "rinha26",
		Concurrency:              256,
		DisableHeaderNamesNormalizing: true,
		NoDefaultServerHeader:    true,
		NoDefaultContentType:     true,
		NoDefaultDate:            true,
		ReadBufferSize:           2048,
		WriteBufferSize:          256,
		MaxRequestBodySize:       8 << 10,
		ReadTimeout:              2 * time.Second,
		WriteTimeout:             2 * time.Second,
		IdleTimeout:              60 * time.Second,
		ReduceMemoryUsage:        false,
	}

	os.Remove(addr)
	ln, err := net.Listen("unix", addr)
	if err != nil {
		log.Fatalf("listen unix: %v", err)
	}
	os.Chmod(addr, 0666)
	log.Fatal(server.Serve(ln))
}
