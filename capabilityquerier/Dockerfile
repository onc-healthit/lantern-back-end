FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go install ./...

RUN go build cmd/main.go 

CMD ["./main"]