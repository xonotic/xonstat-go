build:
	go build -trimpath

.PHONY: test
test:
	go test -cover -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out

swagger:
	swag init -g cmd/web.go
	mv docs/swagger* web/static/

clean:
	rm coverage.out
	rm xonstat
