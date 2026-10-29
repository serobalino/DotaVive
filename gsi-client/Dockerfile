FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd

RUN go mod tidy
RUN go mod download

# Cross-compile Windows + macOS targets; output is mounted at /out.
CMD ["sh", "-c", "set -eu; mkdir -p /out; CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/dota-gsi-windows-amd64.exe ./cmd/client; CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/dota-gsi-darwin-amd64 ./cmd/client; CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o /out/dota-gsi-darwin-arm64 ./cmd/client; echo 'Compilación terminada. Revisa la carpeta output/.'"]
