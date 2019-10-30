FROM golang:1.12

WORKDIR /go/src/app
COPY . .

RUN go install ./...

RUN go build main.go 

CMD ["./main", "./resources/EndpointSources.json"]