// Package vector converts a fraud-score request payload into the 14-dimensional
package vector

// Dim is the fixed dimensionality of the fraud-score vector
const Dim = 14

// TopK is the number of nearest neighbors used for fraud scoring.
const TopK = 5

const Sentinel float64 = -1.0

const DefaultMccRisk float64 = 0.5

type Norm struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio     float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKm                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

type MccRisk map[string]float64
