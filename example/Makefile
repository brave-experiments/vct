binary = vct
godeps = *.go

.PHONY: all
all: lint $(binary)

.PHONY: lint
lint:
	golangci-lint run

$(binary): $(godeps)
	CGO_ENABLED=0 go build -o $(binary)

.PHONY: clean
clean:
	rm -f $(binary)
