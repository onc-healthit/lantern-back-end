FROM golang:1.14
ARG cert_dir

WORKDIR /go/src/app
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .
ENV GO111MODULE=on

RUN go install ./...
RUN go build cmd/main.go 

CMD ["./main"]