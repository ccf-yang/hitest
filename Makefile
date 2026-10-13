# hitest 构建入口（薄封装 scripts/build.sh）
#   make          → 构建全部（Go + 若在 macOS 则含 iOS IPA）
#   make build_go → 只构建 Go 二进制 build/hitest
#   make build_ios→ 只构建未签名 iOS IPA（需 macOS + Xcode）
.PHONY: all build build_go build_ios clean

all: build

build:
	scripts/build.sh all

build_go:
	scripts/build.sh go

build_ios:
	scripts/build.sh ios

clean:
	rm -rf build
