package dto

type CreateShortURLRequest struct {
	LongURL     string `json:"long_url" binding:"required,url"`
	ExpiresDays int    `json:"expires_days" binding:"omitempty,min=0,max=30"`
	Password    string `json:"password" binding:"omitempty,min=6,max=20"`
	Alias       string `json:"alias" binding:"omitempty,min=6,max=20"`
}
