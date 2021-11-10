FROM golang:1.15 AS build-env
WORKDIR /src/dockerhub-limit-exporter-go
ADD go.* /src/dockerhub-limit-exporter-go/
RUN go mod download
ADD . /src/dockerhub-limit-exporter-go/
ENV CGO_ENABLED=0
RUN go build

FROM scratch
COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /src/dockerhub-limit-exporter-go/dockerhub-limit-exporter-go /bin/dockerhub-limit-exporter-go
ENTRYPOINT ["/bin/dockerhub-limit-exporter-go"]