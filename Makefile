.PHONY: test test-integration test-all vet build up down

BACKEND := backend
COMPOSE_TEST := docker compose -f docker-compose.test.yml

vet:
	cd $(BACKEND) && go vet ./...

test:
	cd $(BACKEND) && go test -race ./...

# Integration tests run against the databases from docker-compose.test.yml.
# Packages run one at a time, they share the same databases.
test-integration:
	$(COMPOSE_TEST) up -d --wait
	cd $(BACKEND) && go test -tags integration -race -p 1 -count 1 ./... ; \
		status=$$? ; cd .. ; $(COMPOSE_TEST) down -v ; exit $$status

test-all: vet test test-integration

build:
	cd $(BACKEND) && go build ./...

up:
	docker compose up --build

down:
	docker compose down
