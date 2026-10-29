FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY icons ./icons
COPY macos ./macos
COPY build.sh ./build.sh

RUN go mod tidy
RUN go mod download
RUN go install github.com/tc-hib/go-winres@v0.3.3

# Cross-compile Windows and macOS targets; output is mounted at /out.
CMD ["sh", "./build.sh"]
