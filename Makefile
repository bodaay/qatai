export CGO_ENABLED = 0
export NEXT_TELEMETRY_DISABLED = 1


.PHONY: run
run:
	go run ./cmd/qatai

.PHONY: run-verbose
run-verbose:
	go run ./cmd/qatai --verbose

.PHONY: run-json
run-json:
	go run ./cmd/qatai --json

.PHONY: build
build: build-web build-qatai

.PHONY: build-qatai
build-qatai:
	GOOS=linux GOARCH=amd64 go build -o out/qatai_linux_amd64 ./cmd/qatai
	GOOS=linux GOARCH=arm64 go build -o out/qatai_linux_arm64 ./cmd/qatai
	GOOS=windows GOARCH=amd64 go build -o out/qatai_windows_amd64.exe ./cmd/qatai
	GOOS=darwin GOARCH=amd64 go build -o out/qatai_darwin_x86_64 ./cmd/qatai
	GOOS=darwin GOARCH=arm64 go build -o out/qatai_darwin_arm64 ./cmd/qatai

.PHONY: build-web
build-web:
	cd web && \
	yarn install --frozen-lockfile && \
	yarn run export && \
	rm -rf ../cmd/qatai/web/ \
    mv dist ../cmd/qatai/web

.PHONY: clean
clean:
	rm -f out/*
	rm -rf ./cmd/qatai/web
	rm -rf ./web/dist/
	rm -rf ./web/.next