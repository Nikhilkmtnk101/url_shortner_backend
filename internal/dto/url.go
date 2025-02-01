package dto

type CreateShortURLRequest struct {
	LongURL     string `json:"long_url" binding:"required,url"`
	ExpiresDays int    `json:"expires_days" binding:"omitempty,min=0,max=30"`
}
