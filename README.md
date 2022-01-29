# tempest_influx

Go program to convert WeatherFlow Tempest WX UDP broadcasts to influx
wire protocol.

The [Tempest Weather
System](https://shop.weatherflow.com/products/tempest) sends UDP
broadcasts with weather data and system status periodically.
This program receives those broadcasts and generates InfluxDB wire
protocol messages to import the data into InfluxDB.

This program will forward selected data to InfluxDB or a Telegraf
proxy. XXX - links

## Tempest WX Broadcast Formats

The format of the UDP broadcasts are documented
[here](https://weatherflow.github.io/Tempest/api/udp.html).

The reports of interest are the `obs_st` message which has full
weather data and the `rapid_wind` has instantaneous Wind data.
The later is generated every few seconds and the former once a minute.

## TODO

 + [ ] Pass parameters to the container in envrionment (i.e. token)
   + Use viper for configuration?
 + [ ] Support `bucket_tag` to allow sending to multiple buckets
 + [ ] Optionally send `rapid_wind` data
   + i.e. send rapid_wind to a `daily` bucket with short retention

## Example

The output is designed to be passed to Telegraf for forwarding to
InfluxDB via the influxdb_v2_listener, for example:

```
[[inputs.influxdb_v2_listener]]
  service_address = ":50222"
  bucket_tag = "metrics"
  tls_cert = "/etc/telegraf/keys/cert.pem"
  tls_key = "/etc/telegraf/keys/key.pem"
  token = "SOMEARBITRARYSTRING"
```

## Credits

Inspired by the code in [udpproxy](https://github.com/Akagi201/udpproxy)
