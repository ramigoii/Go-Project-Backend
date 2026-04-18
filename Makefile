POSTGRESQL_URL=postgres://postgres:postgres@localhost:5432/movie_db?sslmode=disable

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

.PHONY: migrate-up
migrate-up:
	migrate -database ${POSTGRESQL_URL} -path migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database ${POSTGRESQL_URL} -path migrations down

.PHONY: migrate-force
migrate-force:
	@read -p "Enter version: " version; \
	migrate -database ${POSTGRESQL_URL} -path migrations force $$version