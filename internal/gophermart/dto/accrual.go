package dto

type AccrualOrderRequest struct {
	Order string        `json:"order"`
	Goods []AccrualGood `json:"goods"`
}

type AccrualGood struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
