FROM golang:1.16
ARG cert_dir

WORKDIR /go/src/app
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
RUN apt-get update
RUN apt-get install -y jq
COPY . .

ENV GO111MODULE=on

RUN go install ./...
RUN go build cmd/sendendpoints/main.go 

CMD ["./main"]