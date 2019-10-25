FROM golang:1.12

WORKDIR /go/src/app
COPY . .

# TODO: Implement Go Modules within project so dependencies can be cached
#RUN go get -d -v ./...

RUN go install -v ./...

RUN go build main.go 

CMD ["./main", "./resources/EndpointSources.json"]