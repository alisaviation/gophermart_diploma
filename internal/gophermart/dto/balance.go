package dto

type WithdrawRequest struct {
	Order string  `json:"order" validate:"required"`
	Sum   float64 `json:"sum" validate:"required,gt=0"`
}

type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
