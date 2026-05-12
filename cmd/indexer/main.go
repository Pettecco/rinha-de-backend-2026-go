package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"log"
	"os"
	"rinha26/internal/consts"
	"rinha26/internal/ivf"
	"time"
)

type Reference struct {
	Vector []float64 `json:"vector"`
	Label  string    `json:"label"`
}

func main() {
	inputPath := flag.String("input", "dataset/references.json.gz", "path to references.json.gz")
	outputPath := flag.String("output", "dataset/ivf.bin", "path to output ivf.bin")
	seed := flag.Uint64("seed", consts.KMeansSeed, "k-means seed")
	limit := flag.Int("limit", 0, "max vectors to load (0 = all)")
	flag.Parse()

	t0 := time.Now()
	log.Printf("loading %s ...", *inputPath)

	vectors, labels, err := loadReferences(*inputPath, *limit)
	if err != nil {
		log.Fatalf("load: %v", err)
	}

	log.Printf("loaded %d vectors in %.1fs", len(vectors), time.Since(t0).Seconds())

	rng := ivf.NewLCG(*seed)
	log.Printf("building IVF index (K=%d, seed=0x%x) ...", consts.K, *seed)

	t1 := time.Now()
	cb, centroids, err := ivf.BuildIndex(vectors, labels, consts.K, rng)
	if err != nil {
		log.Fatalf("build: %v", err)
	}

	log.Printf("index built in %.1fs", time.Since(t1).Seconds())

	f, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("create: %v", err)
	}

	err = ivf.WriteIndex(f, uint32(len(vectors)), centroids, cb)
	f.Close()
	if err != nil {
		log.Fatalf("write: %v", err)
	}

	stat, _ := os.Stat(*outputPath)
	log.Printf("wrote %s (%.1f MB)", *outputPath, float64(stat.Size())/1e6)
	log.Printf("total: %.1fs", time.Since(t0).Seconds())
}

func loadReferences(path string, limit int) ([][]float64, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, err
	}
	defer gz.Close()

	var refs []Reference
	if err := json.NewDecoder(gz).Decode(&refs); err != nil {
		return nil, nil, err
	}

	if limit > 0 && limit < len(refs) {
		refs = refs[:limit]
	}

	vectors := make([][]float64, len(refs))
	labels := make([]byte, len(refs))
	for i, r := range refs {
		vectors[i] = r.Vector
		if r.Label == "fraud" {
			labels[i] = 1
		}
	}

	return vectors, labels, nil
}
