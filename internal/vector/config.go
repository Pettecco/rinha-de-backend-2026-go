package vector

import (
	"encoding/json"
	"os"
)

// LoadNorm reads normalization.json.
func LoadNorm(path string) (*Norm, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	norm := &Norm{}
	if err := json.Unmarshal(raw, norm); err != nil {
		return nil, err
	}
	return norm, nil
}

// LoadMccRisk reads mcc_risk.json into a MccRisk map.
func LoadMccRisk(path string) (MccRisk, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed := map[string]float64{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	out := make(MccRisk, len(parsed))
	for code, risk := range parsed {
		out[code] = risk
	}
	return out, nil
}
