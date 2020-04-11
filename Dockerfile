# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13.5-alpine AS builder

WORKDIR /go/src/quotebot
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir /quotebot

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/quotebot github.com/yfedoruck/quotebot

FROM alpine:3.11
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/quotebot /go/src/quotebot
COPY --from=builder /bin/quotebot /bin/quotebot

# Run the outyet command by default when the container starts.
CMD ["/bin/quotebot"]
EXPOSE 5000