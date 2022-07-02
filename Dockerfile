FROM golang:alpine as builder

# Download and install dependencies
WORKDIR $GOPATH/src/github.com/maesoser/tplink_exporter
COPY . .
RUN adduser -D tplink && go mod tidy

# Compile it
ENV CGO_ENABLED=0
RUN GOOS=linux go build -o tplinkd -a -installsuffix cgo -ldflags '-s -w -extldflags "-static"' .

# Create docker
FROM scratch
COPY --from=builder /go/src/github.com/maesoser/tplink_exporter/tplinkd /app/
USER tplink
ENTRYPOINT ["/app/tplinkd"]
