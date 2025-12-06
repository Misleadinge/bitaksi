# Quick Start Guide

## Prerequisites
- Docker and Docker Compose installed
- Go 1.21+ (if running locally)

## Start Everything with Docker

```bash
# 1. Copy environment file
cp env.example .env

# 2. Start all services
docker-compose up --build

# Services will be available at:
# - Gateway: http://localhost:8080
# - Driver Service: http://localhost:8081
# - MongoDB: localhost:27017
```

## Test the API

### 1. Get JWT Token
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

### 2. Create a Driver
```bash
curl -X POST http://localhost:8080/drivers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "firstName": "Ahmet",
    "lastName": "Demir",
    "plate": "34ABC123",
    "taksiType": "sari",
    "carBrand": "Toyota",
    "carModel": "Corolla",
    "lat": 41.0431,
    "lon": 29.0099
  }'
```

### 3. List Drivers
```bash
curl http://localhost:8080/drivers?page=1&pageSize=20
```

### 4. Find Nearby Drivers
```bash
curl "http://localhost:8080/drivers/nearby?lat=41.0082&lon=28.9784&taksiType=sari"
```

## View Swagger Documentation

- Gateway: http://localhost:8080/swagger/index.html
- Driver Service: http://localhost:8080/swagger/index.html

## Stop Services

```bash
docker-compose down
```

## Run Tests

```bash
make test
```

## Common Issues

### MongoDB Connection Error
- Ensure MongoDB is running: `docker-compose ps`
- Check MongoDB logs: `docker-compose logs mongodb`

### Port Already in Use
- Change ports in `.env` file
- Or stop conflicting services

### JWT Authentication
- If JWT is enabled, you must login first to get a token
- Set `JWT_ENABLED=false` in `.env` to disable authentication for testing

