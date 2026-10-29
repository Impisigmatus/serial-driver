PROJECT=$(shell cat go.mod | grep module | awk -F ' ' '{print $$2}' | awk -F '/' '{print $$NF}')

AUTOGEN=autogen
SPEC=$(AUTOGEN)/oapi3.yaml
TEST=$(AUTOGEN)/tests
REPORTS=.reports
VENDOR=vendor

help: ## Показать описание основных команд
	@echo "\033[1;34m$(PROJECT) usage:\033[0m"
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: dependencies ## Подготовить каталоги для сгенерированных файлов
	@echo -n "\033[1;34mSetup...\033[0m\t"
	@mkdir -p $(AUTOGEN)/server $(TEST)
	@echo "\033[0;32mOK\033[0m"

clean: ## Удалить сгенерированные файлы
	@echo -n "\033[1;34mClean...\033[0m\t"
	@rm -rf $(VENDOR)
	@rm -rf $(REPORTS)
	@rm -rf $(AUTOGEN)
	@echo "\033[0;32mOK\033[0m"

update: setup ## Обновить go.mod и каталог vendor
	@echo -n "\033[1;34mUpdate...\033[0m\t"
	@go mod tidy
	@go mod vendor
	@echo "\033[0;32mOK\033[0m"

format: setup ## Исправить замечания линтера в internal
	@echo -n "\033[1;34mFormat...\033[0m\t"
	@golangci-lint run --fix ./internal/...
	@echo "\033[0;32mOK\033[0m"

lint: format ## Запустить линтер и сохранить отчёт
	@echo -n "\033[1;34mLinter...\033[0m\t"
	@mkdir -p $(REPORTS)
	@golangci-lint run ./internal/... --out-format checkstyle > $(REPORTS)/golangci-lint.xml
	@echo "\033[0;32mOK\033[0m"

mock-gen: setup ## Выполнить директивы go generate
	@echo -n "\033[1;34mMocks...\033[0m\t"
	@go generate ./...
	@echo "\033[0;32mOK\033[0m"

test: gen ## Запустить тесты проекта
	@echo -n "\033[1;34mTest...\033[0m\n"
	@go test ./...
	@echo "\033[0;32mOK\033[0m"

test-cover: gen ## Создать отчёт о покрытии тестами
	@echo -n "\033[1;34mTest coverage...\033[0m\n"
	@go test -coverprofile=$(TEST)/coverage.out ./internal/...
	@go tool cover -html=$(TEST)/coverage.out -o $(TEST)/coverage.html
	@echo "Coverage report generated in $(TEST)/coverage.html"
	@echo "\033[0;32mOK\033[0m"

gen: oapi3-code-gen ## Сгенерировать OpenAPI 3.0 и код сервера

oapi3-code-gen: oapi3-spec-gen ## Сгенерировать код сервера из OpenAPI 3.0
	@echo -n "\033[1;34mCodegen...\033[0m\t"
	@oapi-codegen --config=codegen.yml $(SPEC)
	@echo "\033[0;32mOK\033[0m"

oapi3-spec-gen: update ## Сгенерировать OpenAPI 3.0 из комментариев Go и README
	@echo -n "\033[1;34mSpec...\033[0m\t\t"
	@set -eu; source_dir=$$(mktemp -d "$(AUTOGEN)/.swagger2.XXXXXX"); \
	output=$$(mktemp "$(AUTOGEN)/.oapi3.XXXXXX"); \
	trap 'rm -rf -- "$$source_dir"; rm -f -- "$$output"' EXIT; \
	go run github.com/swaggo/swag/cmd/swag@v1.8.1 init \
		--parseDependency --dir cmd/service,internal/transport \
		--generalInfo main.go --output "$$source_dir" --outputTypes yaml; \
	go run ./cmd/oapi-converter "$$source_dir/swagger.yaml" README.md > "$$output"; \
	chmod 644 "$$output"; \
	mv -f -- "$$output" "$(SPEC)"
	@echo "\033[0;32mOK\033[0m"

dependencies: ## Установить oapi-codegen при отсутствии
	@echo -n "\033[1;34mDependencies...\033[0m\t"
	@command -v oapi-codegen > /dev/null || go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
	@echo "\033[0;32mOK\033[0m"

all: test test-cover ## Выполнить генерацию и проверки

.PHONY: help setup clean update format lint mock-gen test test-cover gen oapi3-code-gen oapi3-spec-gen dependencies all
.DEFAULT_GOAL=all
