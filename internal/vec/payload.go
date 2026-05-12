package vec

import (
	"github.com/buger/jsonparser"
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func FromPayload(raw []byte, dst *[Dim]float64, norm *Norm, mccRisk map[string]float64) error {
	var (
		amount         float64
		installments   float64
		requestedAt    []byte
		custAvgAmount  float64
		custTxCount24h float64
		merchantID     []byte
		merchantMCC    []byte
		merchantAvg    float64
		isOnline       bool
		cardPresent    bool
		kmFromHome     float64
		lastTxTS       []byte
		lastTxKM       float64
		hasLastTx      bool
	)

	var knownMerchants [][]byte

	keys := [][]string{
		{"transaction", "amount"},
		{"transaction", "installments"},
		{"transaction", "requested_at"},
		{"customer", "avg_amount"},
		{"customer", "tx_count_24h"},
		{"customer", "known_merchants"},
		{"merchant", "id"},
		{"merchant", "mcc"},
		{"merchant", "avg_amount"},
		{"terminal", "is_online"},
		{"terminal", "card_present"},
		{"terminal", "km_from_home"},
		{"last_transaction", "timestamp"},
		{"last_transaction", "km_from_current"},
	}

	jsonparser.EachKey(raw, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			amount, _ = jsonparser.ParseFloat(value)
		case 1:
			installments, _ = jsonparser.ParseFloat(value)
		case 2:
			requestedAt = value
		case 3:
			custAvgAmount, _ = jsonparser.ParseFloat(value)
		case 4:
			custTxCount24h, _ = jsonparser.ParseFloat(value)
		case 5:
			jsonparser.ArrayEach(value, func(v []byte, t jsonparser.ValueType, offset int, e error) {
				knownMerchants = append(knownMerchants, append([]byte(nil), v...))
			})
		case 6:
			merchantID = value
		case 7:
			merchantMCC = value
		case 8:
			merchantAvg, _ = jsonparser.ParseFloat(value)
		case 9:
			isOnline = vt == jsonparser.Boolean && (value[0] == 't' || value[0] == 'T')
		case 10:
			cardPresent = vt == jsonparser.Boolean && (value[0] == 't' || value[0] == 'T')
		case 11:
			kmFromHome, _ = jsonparser.ParseFloat(value)
		case 12:
			if vt != jsonparser.Null {
				lastTxTS = value
				hasLastTx = true
			}
		case 13:
			if hasLastTx {
				lastTxKM, _ = jsonparser.ParseFloat(value)
			}
		}
	}, keys...)

	dst[0] = clamp(amount/norm.MaxAmount, 0, 1)
	dst[1] = clamp(installments/norm.MaxInstallments, 0, 1)

	if custAvgAmount > 0 {
		dst[2] = clamp((amount/custAvgAmount)/norm.AmountVsAvgRatio, 0, 1)
	} else {
		dst[2] = 1.0
	}

	dst[3] = float64(ExtractHour(string(requestedAt))) / 23.0
	dst[4] = float64(ExtractDayOfWeek(string(requestedAt))) / 6.0

	if hasLastTx {
		mins := MinutesBetween(string(requestedAt), string(lastTxTS))
		dst[5] = clamp(mins/norm.MaxMinutes, 0, 1)
		dst[6] = clamp(lastTxKM/norm.MaxKm, 0, 1)
	} else {
		dst[5] = Sentinel
		dst[6] = Sentinel
	}

	dst[7] = clamp(kmFromHome/norm.MaxKm, 0, 1)
	dst[8] = clamp(custTxCount24h/norm.MaxTxCount24h, 0, 1)

	if isOnline {
		dst[9] = 1
	} else {
		dst[9] = 0
	}

	if cardPresent {
		dst[10] = 1
	} else {
		dst[10] = 0
	}

	unknown := 1.0
	for _, km := range knownMerchants {
		if string(km) == string(merchantID) {
			unknown = 0.0
			break
		}
	}
	dst[11] = unknown

	risk, ok := mccRisk[string(merchantMCC)]
	if !ok {
		risk = 0.5
	}
	dst[12] = risk

	dst[13] = clamp(merchantAvg/norm.MaxMerchantAvgAmount, 0, 1)

	return nil
}
