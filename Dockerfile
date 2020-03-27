FROM golang:alpine as builder

# Download and install dependencies
RUN apk update && apk add --no-cache git
RUN go get github.com/prometheus/client_golang/prometheus

# Copy the code from the host
WORKDIR $GOPATH/src/github.com/maesoser/tplink_exporter
COPY . .

# Compile it
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' .

# Create docker
FROM scratch
COPY --from=builder /go/src/github.com/maesoser/tplink_exporter/tplink_exporter /app/
ENTRYPOINT ["/app/tplink_exporter"]
