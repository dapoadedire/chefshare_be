I API

ChefShare is a backend application for sharing recipes between chefs and food enthusiasts.

## Features

- User authentication and authorization
- Recipe creation, retrieval, updating, and deletion
- Recipe categorization and searching
- User profiles and favorites
- RESTful API endpoints

## Tech Stack

- Go
- Gin web framework
- PostgreSQL database
- JWT authentication
- Docker containerization

## Getting Started

### Prerequisites

- Go 1.16+
- PostgreSQL
- Docker (optional)

### Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/chefshare_be.git
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

4. Run the application
```bash
go run main.go
```

## API Documentation

API endpoints are available at `/api/v1`

### Authentication
- POST `/api/v1/auth/register` - Register a new user
- POST `/api/v1/auth/login` - Login and get JWT token

### Recipes
- GET `/api/v1/recipes` - List all recipes
- GET `/api/v1/recipes/:id` - Get a specific recipe
- POST `/api/v1/recipes` - Create a new recipe
- PUT `/api/v1/recipes/:id` - Update a recipe
- DELETE `/api/v1/recipes/:id` - Delete a recipe

## License

MIT