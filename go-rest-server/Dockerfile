FROM golang:latest

WORKDIR /go/src/app
COPY . .

RUN go get ./...
RUN go build -o serve
CMD ["./serve"]