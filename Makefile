.PHONY: build-all deploy-all migrate-all seed

build-backend:
	cd apps/garagebrain/backend && go build -o bin/server ./cmd/server

build-gateway:
	cd apps/gateway && go build -o bin/gateway ./cmd/server

build-frontend:
	cd apps/garagebrain/frontend && npm run build

build: build-backend build-gateway build-frontend

build-all: build

dev-backend:
	cd apps/garagebrain/backend && go run ./cmd/server

dev-gateway:
	cd apps/gateway && go run ./cmd/server

dev-frontend:
	cd apps/garagebrain/frontend && npm run dev

dev:
	make -j3 dev-backend dev-gateway dev-frontend

install-deps:
	cd apps/garagebrain/frontend && npm install
	cd apps/garagebrain/backend && go mod download
	cd apps/gateway && go mod download

migrate:
	psql $(DATABASE_URL) -f shared/db/garagebrain_schema.sql

seed:
	cd apps/garagebrain/backend && go run ../../shared/scripts/seed.go

deploy: build-all
	./deploy/deploy.sh

deploy-all: deploy

lint:
	cd apps/garagebrain/backend && go vet ./...
	cd apps/gateway && go vet ./...
	cd apps/garagebrain/frontend && npx eslint src/
