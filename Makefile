generate-swagger-docs:
	$(HOME)/go/bin/swag init

run:
	@if [ -z "$$(docker ps -q -f name=chefshare_be)" ]; then \
		docker compose up -d; \
	fi
	@echo "Running the application..."
	@go run main.go

docker-down:
	@echo "Stopping and removing Docker containers..."
	@docker compose down -v
docker-up:
	@echo "Starting Docker containers..."
	@docker compose up -d