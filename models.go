package main

type CreateShortUrlRequest struct {
	URL string `json:"url"`
}

type UpdateShortUrlRequest struct {
	URL string `json:"url"`
}

type ShortURLResponse struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	ShortCode string `json:"shortCode"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type URLStatsResponse struct {
	ID          int64  `json:"id"`
	URL         string `json:"url"`
	ShortCode   string `json:"shortCode"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	AccessCount int64  `json:"accessCount"`
}
