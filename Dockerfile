FROM golang:alpine AS build

WORKDIR /src/
COPY *.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/tempest_influx

FROM scratch
COPY --from=build /bin/tempest_influx /bin/tempest_influx

EXPOSE 50222/udp

ENTRYPOINT ["/bin/tempest_influx"]

