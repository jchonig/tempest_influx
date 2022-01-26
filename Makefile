IMAGE=tempest_influx:latest
TOKEN=SOMEARBITRARYSTRING
TARGET=https://metrics.home.honig.net:50222/api/v2/write

all: run

run: build
	docker run --net=host -- ${IMAGE} -v --token ${TOKEN} --target ${TARGET}

build:
	docker build -t ${IMAGE} .

compile:
	go build -o influx_tempest .
