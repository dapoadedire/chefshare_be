generate-swagger-docs:
	$HOME/go/bin/swag init

run:
	@if [ -z "$$(docker ps -q -f name=chefshare_be)" ]; then \
		docker compose up -d; \
	fi
	@echo "Running the application..."
	@go run main.go
