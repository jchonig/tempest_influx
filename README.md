# tempest_influx

Go program to convert WeatherFlow Tempest WX UDP broadcasts to influx
wire protocol.

The [Tempest Weather
System](https://shop.weatherflow.com/products/tempest) sends UDP
broadcasts with weather data and system status periodically.
This program receives those broadcasts and generates InfluxDB wire
protocol messages to import the data into InfluxDB.

This program will forward selected data to
[InfluxDB](https://www.influxdata.com/) or a
[Telegraf](https://www.influxdata.com/time-series-platform/telegraf/)
proxy.

Docker host networking is required to receive the UDP broadcasts,
inless some type of proxy is used.

## Tempest WX Broadcast Formats

The format of the UDP broadcasts are documented
[here](https://weatherflow.github.io/Tempest/api/udp.html).

The reports of interest are the `obs_st` message which has full
weather data and the `rapid_wind` has instantaneous Wind data.
The later is generated every few seconds and the former once a minute.

## Configuration

There are three ways to pass configuration information:

A optional YAML configuration file may be provided in *XXX/tempest_influx.yml*
which is read at startup.

Environement variables as described in the table below.  These override configuration file data.

Command line flags, also described in the table below.  These override
configuration file adata and environment variables.

|-----------------------------|----------------|-------------------------------|------------------|--------------------------------------|
| Value                       | Config File    | Environment                   | Flag             | Default                              |
|-----------------------------|----------------|-------------------------------|------------------|--------------------------------------|
| Read buffer size            | buffer         | TEMPEST_INFLUX_BUFFER         | --buffer         | 10240                                |
| Listen Address              | listen_address | TEMPEST_INFLUX_LISTEN_ADDRESS | --listen_address | :50222                               |
| InfluxDB write URL          | influx_url     | TEMPEST_INFLUX_INFLUX_URL     | --influx_url     | https://localhost:8086/api/v2/write |
| Influx authentication token | influx_token   | TEMPEST_INFLUX_INFLUX_TOKEN   | --influx_token   |                                      |
| Influx bucket               | influx_bucket  | TEMPEST_INFLUX_INFLUX_BUCKET  | --influx_bucket  |                                      |
| Verbose logging             | verbose        | TEMPEST_INFLUX_VERBOSE        | -v, --verbose    | False (True if Debug set)            |
| Debug logging               | debug          | TEMPEST_INFLUX_DEBUG          | -d, --debug      | False                                |
|-----------------------------|----------------|-------------------------------|------------------|--------------------------------------|

Notes:

   + *influx_token* is required by *InfluxDB* or *Telegraf* to authenticate requests.
   + *influx_bucket* is not required if configured on the receiving end

## TODO

 + [ ] Pass parameters to the container in envrionment (i.e. token)
   + [ ] Use viper for configuration?
 + [X] Support `bucket_tag` to allow sending to multiple buckets
 + [ ] Optionally send `rapid_wind` data
   + i.e. send rapid_wind to a `daily` bucket with short retention
 + [ ] Optionally send `hub_status` data
   + [ ] Allow specification of a bucket

## Examples

### docker-compose.yml

Following is a sample docker-compose file to run this container.

```yaml
version: "3"

services:
  tempest_influx:
	image: "jchonig/tempest_influx:latest"
    net: host
	environment:
	  TEMPEST_INFLUX_INFLUX_URL: "https://metrics.example.com:8086/api/v2/write"
      TEMPEST_INFLUX_INFLUX_TOKEN: "SOMEARBITRARYSTRING"
      TEMPEST_INFLUX_INFLUX_BUCKET: "weather"
    ports:
	  - 50222/udp
```

### Telegraf

The output is designed to be passed to Telegraf for forwarding to
InfluxDB via the influxdb_v2_listener, for example:

```
[[inputs.influxdb_v2_listener]]
  service_address = ":8086"
  bucket_tag = "metrics"
  tls_cert = "/etc/telegraf/keys/cert.pem"
  tls_key = "/etc/telegraf/keys/key.pem"
  token = "SOMEARBITRARYSTRING"
```

If `bucket_tag` is defined in the Telegraf config file it won't be
necessary to pass it to `tempest_influx`.

## Credits

Inspired by the code in [udpproxy](https://github.com/Akagi201/udpproxy)
