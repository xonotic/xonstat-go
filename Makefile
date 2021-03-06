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

css:
	cd web/static/css && cat foundation.css font-awesome.css app.css luma.css > combined.css
	yuicompressor --type css -o web/static/css/xonstat.css web/static/css/combined.css
	rm web/static/css/combined.css

clean:
	rm coverage.out
	rm xonstat
