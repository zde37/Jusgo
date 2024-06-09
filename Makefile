mongodb:
	docker run --name mongodb -p 27017:27017 -d mongo:5.0-focal

run:
	go run cmd/main.go

test: 
	go test -count=1 -v -cover ./...

mockrepo:
	 mockgen -package mockproviders -destination internal/mock/repo.go  github.com/zde37/Jusgo/internal/repository RepositoryProvider

mockservice:
	 mockgen -package mockproviders -destination internal/mock/service.go  github.com/zde37/Jusgo/internal/service ServiceProvider

.PHONY: run test mongodb mockservice mockrepo