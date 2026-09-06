FROM golang:1.16
ARG cert_dir

WORKDIR /go/src/app
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .

RUN apt-get update && apt-get install -y --only-upgrade ca-certificates && rm -rf /var/lib/apt/lists/*

ENV GO111MODULE=on

RUN go install ./...
RUN go build cmd/main.go 

CMD ["./main"]