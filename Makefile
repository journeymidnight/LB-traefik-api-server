default: build

build:
	go build -o lb-app ./src/
image: build
	docker build -t api -f api.docker .
integrate: image
	cd integrate;go test -check.f APISuite
