package main

import (
	"encoding/json"
	"github/RoanYoskyTimane/go-url-shortener/db"
	"log"
	"net/http"
	"net/url"
	"time"
)

// POST /api/v1/urls
func (u URLHandler) createShortUrl(w http.ResponseWriter, r *http.Request) {
	var req CreateShortUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON payload"}`))
		return
	}

	//Validate url format
	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid URL scheme. Must start with http:// or https://"}`))
		return
	}

	// Generate random code
	shortCode, err := generateShortCode(6)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to generate short code"}`))
		return
	}

	// Save the record on to the database
	ctx := r.Context()
	record, err := u.Queries.CreateURL(ctx, db.CreateURLParams{
		ShortCode:   shortCode,
		OriginalUrl: req.URL,
	})
	if err != nil {
		log.Printf("create URL: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to save URL to database"}`))
		return
	}

	//Store redis cache
	_ = u.RDB.Set(ctx, record.ShortCode, record.OriginalUrl, 24*time.Hour).Err()

	// Build HTTP Response DTO
	resp := ShortURLResponse{
		ID:        record.ID,
		URL:       record.OriginalUrl,
		ShortCode: record.ShortCode,
		CreatedAt: record.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Time.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
func (u URLHandler) getShortUrl(w http.ResponseWriter, r *http.Request)    {}
func (u URLHandler) updateShortUrl(w http.ResponseWriter, r *http.Request) {}
func (u URLHandler) deleteShortUrl(w http.ResponseWriter, r *http.Request) {}
func (u URLHandler) urlStatistics(w http.ResponseWriter, r *http.Request)  {}
