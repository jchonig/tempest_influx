FROM golang:alpine AS build

WORKDIR /src/
COPY *.go go.* /src/
RUN apk --no-cache add ca-certificates; \
     CGO_ENABLED=0 go build -o /bin/tempest_influx

FROM scratch
COPY --from=build /bin/tempest_influx /bin/tempest_influx
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 50222/udp

VOLUME "/config"

ENTRYPOINT ["/bin/tempest_influx"]
