# API Test Script

## Prerequisites
Make sure you have a `.env` file with database configuration:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=your_database
DB_SSLMODE=disable
```

## Test Endpoints

### 1. Health Check
```bash
curl -X GET http://localhost:8080/api/v1/health
```

### 2. Get Recipes (Public)
```bash
curl -X GET "http://localhost:8080/api/v1/recipes?page=1&limit=10"
```

### 3. Get Recipes with Filtering
```bash
curl -X GET "http://localhost:8080/api/v1/recipes?category=dessert&difficulty=easy&sort_by=created_at&sort_order=desc"
```

### 4. Get Single Recipe
```bash
curl -X GET http://localhost:8080/api/v1/recipes/1
```

### 5. Get User's Recipes
```bash
curl -X GET http://localhost:8080/api/v1/users/testuser/recipes
```

### 6. Create Recipe (Requires Authentication)
```bash
curl -X POST http://localhost:8080/api/v1/recipes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Test Recipe",
    "description": "A delicious test recipe",
    "difficulty_level": "easy",
    "status": "published",
    "prep_time": 15,
    "cook_time": 30,
    "total_time": 45,
    "ingredients": [
      {
        "name": "Sugar",
        "quantity": 2,
        "unit": "cups",
        "position": 1
      }
    ],
    "steps": [
      {
        "step_number": 1,
        "instruction": "Mix all ingredients together"
      }
    ]
  }'
```

### 7. Update Recipe (Requires Authentication + Ownership)
```bash
curl -X PUT http://localhost:8080/api/v1/recipes/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Updated Test Recipe",
    "description": "An updated description"
  }'
```

### 8. Delete Recipe (Requires Authentication + Ownership)
```bash
curl -X DELETE http://localhost:8080/api/v1/recipes/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 9. Access Swagger Documentation
Open browser to: http://localhost:8080/swagger/index.html

## Expected Responses

- All public endpoints should return 200 OK
- Protected endpoints without auth should return 401 Unauthorized
- Trying to modify someone else's recipe should return 403 Forbidden
- Invalid recipe IDs should return 400 Bad Request
- Non-existent recipes should return 404 Not Found