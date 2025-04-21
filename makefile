run:
	go run cmd/main.go
build:
	go build -o bin/sndbx ./cmd/main.go
test:
	go test -v ./internal
dev-test:
	cd test && DOCKER_API_VERSION=1.48 ../bin/sndbx init
dev-reset-env:
	docker image rm -f sndbx-test
	docker rm -f sndbx-test
install:
	go build -o bin/sndbx ./cmd/main.go
	@echo "sndbx build to bin/sndbx"
	sudo cp bin/sndbx /bin/sndbx
	@echo "sndbx installed to /bin/sndbx"
uninstall:
	sudo rm -f /bin/sndbx
	@echo "sndbx uninstalled from /bin/sndbx"