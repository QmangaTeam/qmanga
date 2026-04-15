.PHONY: up down build-backend build-frontend bootstrap
up:
	docker-compose up --build -d

down:
	docker-compose down

build-backend:
	docker build -t qmanga-backend ./backend

bootstrap:
	bash ./scripts/bootstrap.sh

migrate:
	docker-compose run --rm api migrate

local-migrate:
	cd backend && go run main.go migrate
