# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13.5-alpine
#FROM golang:alpine
RUN apk add --no-cache git

RUN mkdir /quotebot

# Copy the local package files to the container's workspace.
ADD . /go/src/quotebot

RUN go get -u gopkg.in/telegram-bot-api.v4
RUN go install quotebot

# Run the outyet command by default when the container starts.
CMD /go/bin/quotebot
