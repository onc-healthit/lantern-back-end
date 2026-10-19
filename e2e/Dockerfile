FROM golang:1.26.9-trixie
ARG cert_dir

WORKDIR /go/src/github.com/onc-healthit/lantern-back-end/e2e
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .


ENV GO111MODULE=on

# RUN go install ./...
