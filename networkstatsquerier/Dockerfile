FROM golang:1.14

WORKDIR /go/src/app
COPY . .

RUN go install ./...

RUN go build main.go 

CMD ["./main", "./resources/EndpointSources.json", "CareEvolution"]