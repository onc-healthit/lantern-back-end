FROM golang:1.16
ARG cert_dir
ARG force_version
ARG direction
ENV env_force_version=$force_version
ENV env_direction=$direction

WORKDIR /go/src/app
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .
ENV GO111MODULE=on

RUN go install ./...
RUN go build cmd/main.go 

CMD ./main $env_direction $env_force_version
