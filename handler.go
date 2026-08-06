package main

import (
	"context"
	"encoding/json"
	"github/RoanYoskyTimane/go-url-shortener/db"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
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

// GET /{shortCode}
func (u URLHandler) getShortUrl(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
		return
	}

	ctx := r.Context()

	// Check Redis cache
	targetURL, err := u.RDB.Get(ctx, shortCode).Result()
	if err == nil {
		go func() {
			_ = u.Queries.IncrementAccessCount(context.Background(), shortCode)
		}()

		http.Redirect(w, r, targetURL, http.StatusFound)
		return
	}

	record, err := u.Queries.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
		return
	}

	_ = u.RDB.Set(ctx, record.ShortCode, record.OriginalUrl, 24*time.Hour).Err()

	// Increment access count in database
	go func() {
		_ = u.Queries.IncrementAccessCount(context.Background(), record.ShortCode)
	}()

	http.Redirect(w, r, record.OriginalUrl, http.StatusFound)
}

func (u URLHandler) updateShortUrl(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
		return
	}

	var req UpdateShortUrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid JSON payload"}`))
		return
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid URL scheme. Must start with http:// or https://"}`))
		return
	}

	ctx := r.Context()
	record, err := u.Queries.UpdateURL(ctx, db.UpdateURLParams{
		OriginalUrl: req.URL,
		ShortCode:   shortCode,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Short URL not found"}`))
		return
	}

	_ = u.RDB.Set(ctx, record.ShortCode, record.OriginalUrl, 24*time.Hour).Err()

	resp := ShortURLResponse{
		ID:        record.ID,
		URL:       record.OriginalUrl,
		ShortCode: record.ShortCode,
		CreatedAt: record.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: record.UpdatedAt.Time.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (u URLHandler) deleteShortUrl(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not found"}`))
		return
	}

	ctx := r.Context()

	err := u.Queries.DeleteURLByShortCode(ctx, shortCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Short URL Not found"}`))
		return
	}
	_ = u.RDB.Del(ctx, shortCode).Err()

	w.WriteHeader(http.StatusNoContent)
}

func (u URLHandler) urlStatistics(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	if shortCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Short code parameter required"}`))
		return
	}

	ctx := r.Context()

	record, err := u.Queries.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Short URL Not found"}`))
		return
	}

	resp := URLStatsResponse{
		ID:          record.ID,
		URL:         record.OriginalUrl,
		ShortCode:   record.ShortCode,
		CreatedAt:   record.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   record.UpdatedAt.Time.Format(time.RFC3339),
		AccessCount: record.AccessCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
