include .env
export

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-down-all:
	migrate -path ./migrations -database "$(DATABASE_URL)" down -all

migrate-drop:
	migrate -path ./migrations -database "$(DATABASE_URL)" drop

migrate-drop-force:
	migrate -path ./migrations -database "$(DATABASE_URL)" drop -f

migrate-version:
	migrate -path ./migrations -database "$(DATABASE_URL)" version

migrate-nuke:
	psql -U postgres -d chat_app -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO postgres;"

migrate-reset: migrate-nuke migrate-up

run:
	go run main.go