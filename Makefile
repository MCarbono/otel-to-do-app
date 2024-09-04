
run:
	go run main.go --env=local

infra_up:
	docker-compose up -d

infra_down:
	docker-compose -f docker-compose.yml down

todos_run:
	go run ./todos/main.go

api_run:
	go run ./api/main.go