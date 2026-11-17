.PHONY: build build-engine build-cli

ifeq ($(OS),Windows_NT)
BINARY := bin/snipbox.exe
ENGINE_TARGET := --target x86_64-pc-windows-gnu
MAKE_BIN := if not exist bin mkdir bin
LINK_FLAGS := -s -w
LINK_FLAGS += -linkmode external -extldflags '-static'
CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER ?= x86_64-w64-mingw32-gcc
export CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER
else
BINARY := bin/snipbox
ENGINE_TARGET :=
MAKE_BIN := mkdir -p bin
LINK_FLAGS := -s -w
ifeq ($(shell go env GOOS),linux)
LINK_FLAGS += -linkmode external -extldflags '-static'
endif
endif

build-engine:
	cargo build --release $(ENGINE_TARGET) --manifest-path engine/Cargo.toml

build-cli: build-engine
	$(MAKE_BIN)
	cd cli && go build -trimpath -ldflags="$(LINK_FLAGS)" -o ../$(BINARY) .

build: build-cli
