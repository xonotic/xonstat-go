build:
	go build -trimpath

docker-build:
	docker build -t xonstat-build .
	docker container create --name xonstat-build xonstat-build
	docker container cp xonstat-build:/home/xonstat/xonstat .
	docker container rm xonstat-build

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
