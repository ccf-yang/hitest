# hitest 构建入口（薄封装 scripts/build.sh）
#   make          → 构建全部（Go + 若在 macOS 则含 iOS IPA）
#   make build_go → 只构建 Go 二进制 build/hitest
#   make build_ios→ 只构建未签名 iOS IPA（需 macOS + Xcode）
.PHONY: all build build_go build_ios test e2e agent-arm64 agent-amd64 clean

all: build

build:
	scripts/build.sh all

build_go:
	scripts/build.sh go

build_ios:
	scripts/build.sh ios

icon:
	scripts/build.sh icon

test:
	go test ./... -timeout 300s

e2e:
	go test ./internal/server/ -run TestE2E -v -timeout 300s

agent-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/hitest-linux-arm64 ./cmd/hitest

agent-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/hitest-linux-amd64 ./cmd/hitest

clean:
	rm -rf bin dist
