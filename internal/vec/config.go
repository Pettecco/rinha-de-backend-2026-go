package vec

import (
	"encoding/json"
	"os"
)

type Norm struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio     float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKm                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

func LoadNorm(path string) (*Norm, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var n Norm
	if err := json.Unmarshal(data, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

func LoadMccRisk(path string) (map[string]float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]float64
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
