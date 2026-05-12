package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"rinha26/internal/consts"
	"rinha26/internal/ivf"
)

type Reference struct {
	Vector []float64 `json:"vector"`
	Label  string    `json:"label"`
}

func main() {
	refs, err := loadReferences("dataset/references.json.gz")
	if err != nil {
		log.Fatalf("load: %v", err)
	}

	rng := ivf.NewLCG(consts.KMeansSeed)
	cb, centroids, err := ivf.BuildIndex(refs.vectors, refs.labels, consts.K, rng)
	if err != nil {
		log.Fatalf("build: %v", err)
	}

	f, err := os.Create("dataset/ivf.bin")
	if err != nil {
		log.Fatalf("create: %v", err)
	}

	if err := ivf.WriteIndex(f, uint32(len(refs.vectors)), centroids, cb); err != nil {
		f.Close()
		log.Fatalf("write: %v", err)
	}

	f.Close()
}

type refs struct {
	vectors [][]float64
	labels  []byte
}

func loadReferences(path string) (*refs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var items []Reference
	if err := json.NewDecoder(gz).Decode(&items); err != nil {
		return nil, err
	}

	r := &refs{
		vectors: make([][]float64, len(items)),
		labels:  make([]byte, len(items)),
	}
	for i, item := range items {
		r.vectors[i] = item.Vector
		if item.Label == "fraud" {
			r.labels[i] = 1
		}
	}

	return r, nil
}
