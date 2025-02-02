# URL Shortener Backend Service

## Overview
This is the backend service for a robust and scalable URL Shortener application. It provides high-performance URL shortening, redirection, and analytics with advanced security features. The service is built using **Golang**, **PostgreSQL**, **Redis**, and deployed on **AWS**.

## Features
- **URL Shortening**: Convert long URLs into short, easily shareable links.
- **Custom Aliases**: Users can create custom short links.
- **Redirection**: Seamless redirection to original URLs.
- **Expiry Dates**: Set expiration dates for short links.
- **Bulk Creation**: Generate multiple short URLs at once.
- **Analytics Dashboard**: Track link clicks, geographic locations, referrers, and more.
- **QR Code Generation**: Generate QR codes for shortened URLs.
- **User Management**: Secure authentication and authorization using JWT.
- **Security Measures**: Spam prevention, password-protected links, and data encryption.
- **Scalability**: Fast lookups, load balancing, and caching with Redis.

## Tech Stack
- **Backend**: Golang
- **Database**: PostgreSQL
- **Cache**: Redis
- **Security**: JWT authentication, encryption
- **Infrastructure**: AWS (S3, EC2, Lambda, RDS)

## Installation
### Prerequisites
- Golang installed
- PostgreSQL database set up
- Redis server running

### Setup
1. Clone the repository:
   ```sh
   git clone https://github.com/your-repo/url-shortener-backend.git
   cd url-shortener-backend
   ```
2. Set up environment variables in a `.env` file:
   ```ini
   ENV = local
   COMPONENT = URL_SHORTNER_SERVICE
   SERVER_PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=url_shortner_backend_app
   DB_PASSWORD=qQHeVnuPjDEm9TI
   DB_NAME=url_shortner_db
   ACCESS_JWT_SECRET=AlphaBeta
   REFRESH_JWT_SECRET=AlphaBeta
    
   REDIS_HOST=localhost
   REDIS_PORT=6379
    
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=yourmail@gmail.com
   SMTP_PASSWORD=yourpassword
   FROM_EMAIL=yourmail@gmail.com

   ```
3. Install dependencies:
   ```sh
   go mod tidy
   ```
4. Run database migrations:
   ```sh
   go run migrate.go
   ```
5. Start the server:
   ```sh
   go run main.go
   ```
