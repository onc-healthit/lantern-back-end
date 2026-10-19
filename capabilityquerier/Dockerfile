FROM golang:1.26.9-trixie
ARG cert_dir

WORKDIR /go/src/app
RUN apt-get update && apt-get install -y --no-install-recommends --only-upgrade ca-certificates && rm -rf /var/lib/apt/lists/*
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .

ENV GO111MODULE=on

RUN go install ./...
RUN go build cmd/main.go 

CMD ["./main"]