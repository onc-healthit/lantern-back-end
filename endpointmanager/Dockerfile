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

# Create a directory for profiling data
RUN mkdir -p /var/lantern/profiles

# Modify the CMD to include profiling flags
# Example:
# CMD ["/etc/lantern/wait-for-it.sh", "lantern-mq:5672", "--", "/etc/lantern/wait-for-it.sh", "postgres:5432", "--", "./main", "-httpprofile=:6060", "-memprofile=/var/lantern/profiles/mem.prof"]

# Expose profiling port
EXPOSE 6060

# Add volume for profiling data
VOLUME ["/var/lantern/profiles"]
