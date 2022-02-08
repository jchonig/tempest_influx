IMAGE=tempest_influx:latest
FLAGS=-v
APP=tempest_influx

all: ${APP}

${APP}: *.go
	go build -o ${APP} .

test: ${APP}
	TEMPEST_INFLUX_CONFIG_DIR=$${PWD}/config ./tempest_influx ${FLAGS}

clean:
	go clean

get:
	go get

fmt:
	go fmt

vet:
	go vet

docker-run: docker-build
	test -f config/tempest_influx.yml && VOLUMES="-v $${PWD}/config:/config" && \
		docker run --net=host $${VOLUMES:-} -- ${IMAGE} ${FLAGS}

docker-build:
	docker build -t ${IMAGE} .

