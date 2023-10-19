lint:
	@echo "$(OK_COLOR)==> Linting with golangci-lint$(NO_COLOR)"
	@docker run --rm -v `pwd`:/app -w /app golangci/golangci-lint:v1.54.2 golangci-lint run -v

test:
	@echo "$(OK_COLOR)==> Running tests using docker-compose deps$(NO_COLOR)"
	@docker compose up -d --wait
	@PG_URI="postgres://test:test@`docker-compose port postgres 5432`/test?sslmode=disable" \
		go test -timeout 60s -cover -coverprofile=coverage.txt -covermode=atomic ./...
