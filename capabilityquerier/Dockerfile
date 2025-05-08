FROM golang:1.16
ARG cert_dir

WORKDIR /go/src/app
COPY ${cert_dir}/ /etc/ssl/certs
RUN update-ca-certificates
COPY . .

ENV GO111MODULE=on

RUN go install ./...

# Build the command properly and put it in the expected location
RUN cd cmd && go build -o /go/src/app/main .

# Expose port for pprof
EXPOSE 6061

CMD ["./main"]