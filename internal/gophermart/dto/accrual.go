package dto

type AccrualOrderRequest struct {
	Order string        `json:"order"`
	Goods []AccrualGood `json:"goods"`
}

type AccrualGood struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Reward      float64 `json:"reward,omitempty"`
	RewardType  string  `json:"reward_type,omitempty"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type AccrualGoodReward struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}
