package service

type AccessToken struct {
	Subject   string  `json:"sub"`
	ExpiresAt float64 `json:"exp"`
	Key       string  `json:"key"`
}
