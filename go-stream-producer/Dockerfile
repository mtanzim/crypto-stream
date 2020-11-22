FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get ./...
RUN go build -o publish
CMD ["./publish"]