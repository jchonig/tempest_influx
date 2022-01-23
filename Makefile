IMAGE=tempest_influx:latest
TOKEN=SOMEARBITRARYSTRING

all: run

run: build
	docker run --net=host -- ${IMAGE} -v -token ${TOKEN}

build:
	docker build -t ${IMAGE} .

compile:
	go build -o influx_tempest .
