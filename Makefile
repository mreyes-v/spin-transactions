.PHONY: test build docker db db-down db-reset local deploy

test:
	go test ./... -count=1

build:
	CGO_ENABLED=0 go build -o dist/write ./cmd/write
	CGO_ENABLED=0 go build -o dist/read ./cmd/read

docker:
	docker build --platform linux/arm64 -t spin-transactions:local .

db:
	docker compose up -d --wait

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d --wait

local:
	bash scripts/local.sh

deploy:
	bash scripts/deploy.sh
