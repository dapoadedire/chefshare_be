# ChefShare API

ChefShare is a backend application for sharing recipes between chefs and food enthusiasts. The platform allows users to create, discover, and interact with recipes in a social cooking community.

**Status: Work In Progress** 🚧

## Features

- **User Management**

  - Registration and authentication with JWT
  - Email verification system
  - Password reset with OTP
  - Profile management

- **Recipe Management**
  - Create, read, update, delete (CRUD) operations
  - Categorization and tagging system
  - Difficulty levels (easy, medium, hard)
  - Detailed recipe information (prep time, cook time, serving size)
- **Social Features**

  - Reviews and ratings
  - Bookmarking favorite recipes
  - Like/unlike recipes

- **API Security**
  - JWT-based authentication
  - Token refresh mechanism
  - Token blacklisting for logout
  - Rate limiting for sensitive operations

## Tech Stack

- **Language**: Go 1.24+
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance web framework
- **Database**: PostgreSQL with migrations using [Goose](https://github.com/pressly/goose)
- **Authentication**: JWT (JSON Web Tokens) with refresh token mechanism
- **Documentation**: [Swagger/OpenAPI](https://github.com/swaggo/gin-swagger)
- **Email Service**: [Resend.com](https://resend.com) for transactional emails
- **Infrastructure**: Docker for containerization

## Project Structure

```
chefshare_be/
├── api/             # HTTP handlers for API endpoints
├── app/             # Application setup and configuration
├── docs/            # Auto-generated Swagger documentation
├── internal/        # Core business logic and domain models
├── middleware/      # HTTP middleware (auth, rate limiting, etc.)
├── migrations/      # Database migrations managed by Goose
├── routes/          # API route definitions
├── services/        # Business service implementations
├── store/           # Database access and repository layer
└── utils/           # Helper utilities and common functions
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Docker and Docker Compose (for local development)

### Installation

1. Clone the repository

```bash
git clone https://github.com/dapoadedire/chefshare_be.git
cd chefshare_be
```

2. Install dependencies

```bash
go mod download
```

3. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start the PostgreSQL database using Docker

```bash
make docker-up
```

5. Run the application

```bash
make run
```

6. Access the API at `http://localhost:8080/api/v1`
7. Access Swagger documentation at `http://localhost:8080/swagger/index.html`

## API Documentation

API endpoints are available at `/api/v1`

### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and get JWT token
- `POST /api/v1/auth/token/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout and invalidate tokens
- `POST /api/v1/auth/verify-email/confirm` - Verify email address
- `POST /api/v1/auth/password/reset/request` - Request password reset

### User Management

- `GET /api/v1/auth/me` - Get authenticated user profile

### Recipes

- `GET /api/v1/recipes` - List all recipes
- `GET /api/v1/recipes/:id` - Get a specific recipe
- `POST /api/v1/recipes` - Create a new recipe
- `PUT /api/v1/recipes/:id` - Update a recipe
- `DELETE /api/v1/recipes/:id` - Delete a recipe

### Health Check

- `GET /api/v1/health` - Check API and database health

## Development

### Generate Swagger Documentation

```bash
make generate-swagger-docs
```

### Manage Docker

```bash
# Start PostgreSQL container
make docker-up

# Stop and remove containers
make docker-down
```

## License

MIT
