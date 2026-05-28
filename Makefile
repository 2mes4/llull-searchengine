.PHONY: build test run seed docker-up docker-down docker-push npm-publish flutter-publish clean lint all

build:
	go build ./...

test:
	go test ./... -v -race -count=1

lint:
	go vet ./...

run: seed.json
	go run ./cmd/server -seed-file seed.json -port 8080

seed.json:
	go run ./cmd/server -generate-seed seed.json -seed-dir data/llibres-llull -seed-count 1000

docker-up:
	docker compose -f deploy/docker-compose.yml up --build -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

docker-push:
	docker build -t llull-searchengine:latest -f deploy/docker/Dockerfile.server .
	docker tag llull-searchengine:latest docker.io/llull-searchengine:latest
	docker tag llull-searchengine:latest ghcr.io/2mes4/llull-searchengine:latest
	@echo "Run: docker push docker.io/llull-searchengine:latest"
	@echo "Run: docker push ghcr.io/2mes4/llull-searchengine:latest"

npm-publish:
	cd ui-components/react && npm publish --access public

flutter-publish:
	cd ui-components/flutter && flutter pub publish

k8s-deploy:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/

clean:
	rm -f seed.json

all: lint build test
