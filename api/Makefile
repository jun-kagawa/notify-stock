.PHONY:db-setup\ 
	build\
	notify\
	gqlgen\
	update\
	update-all\
	gen-local-key\
	db-connect\
	run-api-server\
	deploy\
	fmt\
	test

db-setup:
	docker compose up -d database
	PGPASSWORD=postgres psql -d notify-stock -U postgres -p 5555 -h localhost < schema.sql
	PGPASSWORD=postgres psql -d notify-stock-test -U postgres -p 5555 -h localhost < schema.sql

db-connect:
	psql -d notify-stock -h localhost -U postgres -p 5555

notify:
	/usr/local/go/bin/go run cmd/main.go notify -s "^N225,^GSPC"
update:
	/usr/local/go/bin/go run cmd/main.go stock update
update-all:
	/usr/local/go/bin/go run cmd/main.go stock update -a
gqlgen:
	go run github.com/99designs/gqlgen generate

run-api-server:
	go run cmd/main.go server --tls
deploy:
	docker compose -f compose.prod.yaml up -d --build
fmt:
	go fmt ./...
test: db-setup
	go test ./... -count=1
