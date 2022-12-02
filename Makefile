NAME = lar
GO_BUILD = go build
RELEASE_DIR ?= release
$(shell mkdir -p ${RELEASE_DIR})

all:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC=clang $(GO_BUILD) -o ${RELEASE_DIR}/$(NAME)

darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC=clang $(GO_BUILD) -o ${RELEASE_DIR}/$(NAME)-darwin -v

linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc $(GO_BUILD) -o ${RELEASE_DIR}/$(NAME)-linux -v

windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(GO_BUILD) -o ${RELEASE_DIR}/$(NAME)-linux -v