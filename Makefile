.PHONY: build run test clean migrate

# Переменные
BINARY_NAME=gophermart
BUILD_DIR=bin
MIGRATIONS_DIR=migrations

# Сборка приложения
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gophermart

# Запуск приложения
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Тестирование
test:
	@echo "Running tests..."
	go test -v ./...

# Очистка
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

# Миграции базы данных
migrate:
	@echo "Running database migrations..."
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URI)" up

# Создание миграции
migrate-create:
	@echo "Creating new migration..."
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

# Установка зависимостей
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Линтинг
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

# Форматирование кода
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Проверка кода
check: fmt lint test

# Скачивание statictest
download-statictest:
	@echo "Downloading statictest..."
	./scripts/download_statictest.sh

# Запуск statictest
statictest:
	@echo "Running statictest..."
	go vet -vettool=./.tools/statictest ./...

# Полная проверка кода
check-all: fmt lint test statictest
