LIB_PATH := $(shell pwd)/lib

ifeq ($(shell uname -s),Darwin)
LIB_EXT=dylib
ARCH_OS=osx-amd64
LIBRARY_PATH := DYLD_LIBRARY_PATH=$(LIB_PATH)
else
LIB_EXT=so
ARCH_OS=linux-amd64
LIBRARY_PATH := LD_LIBRARY_PATH=$(LIB_PATH)
endif
LIBS := lib/libduckdb.$(LIB_EXT)
LDFLAGS := LIB=libduckdb.$(LIB_EXT) CGO_LDFLAGS="-L$(LIB_PATH)" $(LIBRARY_PATH) CGO_CFLAGS="-I$(LIB_PATH)"

$(LIBS):
	mkdir -p lib
	curl -Lo lib/libduckdb.zip https://github.com/duckdb/duckdb/releases/download/v0.3.2/libduckdb-$(ARCH_OS).zip
	cd lib; unzip libduckdb.zip

.PHONY: build
build: $(LIBS)
	$(LDFLAGS) go run examples/simple.go

.PHONY: test
test: $(LIBS)
	$(LDFLAGS) go test -v ./...

.PHONY: clean
clean:
	rm -rf lib