# metrika

## Requirements

- Go 1.20+
- [Ginkgo](https://github.com/onsi/ginkgo), [Gomega](https://github.com/onsi/gomega), [GoMock](https://github.com/golang/mock/gomock)

## Setup

1. Install dependencies:
```bash
    go mod tidy 
    go get github.com/onsi/ginkgo/v2 
    go get github.com/onsi/gomega 
    go get github.com/golang/mock/gomock
```

2. Generate mocks: 
#### NOTE: this is not required because I already commited the generated mocks.
    
```bash
    make mocks
```

## Running

Start the service:
```bash
    go run main.go
```
or

```bash
    make docker-up
```

## Testing

Run all tests:
```bash
    make test-ginkgo
```
or

```bash
    go clean -testcache && go test ./... -v
```
