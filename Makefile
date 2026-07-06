.PHONY: test test-go test-web test-e2e lint coverage ci-fast

export PATH := /usr/local/go/bin:$(PATH)

test: test-go test-web test-e2e

test-go:
	go test -count=1 ./...

test-web:
	cd web && npm test

test-e2e:
	go test -count=1 ./e2e/...

lint:
	golangci-lint run --config .github/golangci.yml ./...
	cd web && npm run typecheck

coverage:
	chmod +x scripts/check-coverage.sh
	./scripts/check-coverage.sh

ci-fast: lint test coverage
