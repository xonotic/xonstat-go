build:
	go build

.PHONY: test
test:
	go test -cover -coverprofile=coverage.out ./...

coverage: test
	go tool cover -html=coverage.out

clean:
	rm coverage.out
	rm xonstat
