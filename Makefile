IMAGE=tempest_influx:latest
FLAGS=-v

all: run

run: build
	test -f config/tempest_influx.yml && VOLUMES="-v $${PWD}/config:/config" && \
		docker run --net=host $${VOLUMES:-} -- ${IMAGE} ${FLAGS}

build:
	docker build -t ${IMAGE} .

compile:
	go build -o tempest_influx .
