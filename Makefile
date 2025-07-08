# Makefile

APP_NAME := web-analyzer
PKGS := ./internal/controllers ./internal/services ./internal/web_analyzer_utils
COVERAGE_OUT := coverage.out

test:
	go test -v $(PKGS)

coverage:
	go test -coverprofile=$(COVERAGE_OUT) $(PKGS)
	go tool cover -func=$(COVERAGE_OUT)

coverage-html:
	go test -coverprofile=$(COVERAGE_OUT) $(PKGS)
	go tool cover -html=$(COVERAGE_OUT)