package currencybalance

import (
	"encoding/json"
)

type CurrencyBalance struct {
	Balance int64
}

func (s *CurrencyBalance) GetFloat() float64 {
	return float64(s.Balance) / 100
}

func (s *CurrencyBalance) SetFloat(value float64) {
	s.Balance = int64(value * 100)
}

func (s *CurrencyBalance) IsNegative() bool {
	return s.Balance < 0
}

func (s *CurrencyBalance) Less(other CurrencyBalance) bool {
	return s.Balance < other.Balance
}

func (s *CurrencyBalance) Withdraw(other CurrencyBalance) {
	s.Balance -= other.Balance
}

func (s *CurrencyBalance) Add(other CurrencyBalance) {
	s.Balance += other.Balance
}

func (s CurrencyBalance) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.GetFloat())
}

func (s *CurrencyBalance) UnmarshalJSON(data []byte) error {
	var floatValue float64
	if err := json.Unmarshal(data, &floatValue); err != nil {
		return err
	}
	s.SetFloat(floatValue)
	return nil
}
