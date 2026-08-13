# Link Precision Backend (Go)

## What it is
A high-performance URL shortener REST API backend written in Go, serving as the core engine for "Link Precision".

## What it does
* **URL Shortening**: Generates unique short codes for long web URLs.
* **Redirection Engine**: Redirects incoming short URL requests to the destination URL.
* **Management API**: Allows clients to register, update destination links, and delete short-code records.
* **Click Analytics**: Tracks click counts and statistics for shortened codes.
* **Dual Database Architecture**:
  * **PostgreSQL**: Used for persistent storage of original/shortened URL records. Automatically runs setup schema on startup.
  * **Redis**: Acts as an ephemeral cache for hot redirection codes to enable rapid redirects.

## How to execute it
### Prerequisites
* Go 1.21+ installed on your system.
* Active **PostgreSQL** instance.
* Active **Redis** server.

### Steps
1. **Configure Environment Variables**:
   Create a `.env` file in the root directory:
   ```env
   PORT=3000
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_postgres_password
   DB_NAME=urlShortener
   DB_SSLMODE=disable
   REDIS_ADDR=localhost:6379
   REDIS_PASSWORD=
   ALLOWED_ORIGINS=http://localhost*
   ```

2. **Setup Databases**:
   Ensure PostgreSQL contains a database named `urlShortener` matching your `.env` settings. The backend automatically applies schema tables from `sql/schema.sql` when it connects.

3. **Run Backend**:
   ```bash
   go run .
   ```
   The backend server will start and listen on the configured PORT (defaults to `3000`).

### REST API Endpoints
* **Create Short Code**:
  `POST /api/v1/urls`
  *(Body: JSON mapping destination URL)*
* **Update Target URL**:
  `PUT /api/v1/urls/{shortCode}`
* **Delete Short Code**:
  `DELETE /api/v1/urls/{shortCode}`
* **Retrieve Analytics**:
  `GET /api/v1/urls/{shortCode}/stats`
* **Redirection Endpoint**:
  `GET /{shortCode}`
