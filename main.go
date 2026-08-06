package main

import (
	"context"
	"fmt"
	"github/RoanYoskyTimane/go-url-shortener/db"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type URLHandler struct {
	Queries *db.Queries
	RDB     *redis.Client
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSL := os.Getenv("DB_SSLMODE")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, dbSSL)

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("PostgreSQL ping failed: %v\n", err)
	}
	log.Println("Connected to PostgreSQL successfully")

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v\n", err)
	}
	log.Println("Connected to Redis successfully")

	queries := db.New(dbPool)

	urlHandler := URLHandler{
		Queries: queries,
		RDB:     rdb,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1/urls", func(r chi.Router) {
		r.Post("/", urlHandler.createShortUrl)
		r.Put("/{shortCode}", urlHandler.updateShortUrl)
		r.Delete("/{shortCode}", urlHandler.deleteShortUrl)
		r.Get("/{shortCode}/stats", urlHandler.urlStatistics)
	})

	r.Get("/{shortCode}", urlHandler.getShortUrl)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))


}
