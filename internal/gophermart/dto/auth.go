package dto

type RegisterRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=30"`
	Password string `json:"password" validate:"required,min=6,max=30"`
}
