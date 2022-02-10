package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"

	"github.com/de-wax/go-pkg/dewpoint"
)

var PrecipitationType []string = []string{"none", "rain", "hail", "rain+hail"}

type Report struct {
	StationSerial string       `json:"serial_number,omitempty"`
	ReportType    string       `json:"type"`
	HubSerial     string       `json:"hub_sn,omitempty"`
	Obs           [1][]float64 `json:"obs,omitempty"`
	Ob            [3]float64   `json:"ob,omitempty"`
	FirmwareRevision int
	Uptime       int       `json:"uptime,omitempty"`
	Timestamp    int       `json:"timestamp,omitempty"`
	ResetFlags   string    `json:"reset_flags,omitempty"`
	Seq          int       `json:"seq,omitempty"`
	Fs           []float64 `json:"fs,omitempty"`
	Radio_Stats  []float64 `json:"radio_stats,omitempty"`
	Mqtt_Stats   []float64 `json:"mqtt_stats,omitempty"`
	Voltage      float64   `json:"voltage,omitempty"`
	RSSI         float64   `json:"rssi,omitempty"`
	HubRSSI      float64   `json:"hub_rssi,omitempty"`
	SensorStatus int       `json:"sensor_status,omitempty"`
	Debug        int       `json:"debug,omitempty"`
}

func tempest_obs_st(report Report, m *InfluxData) {
	type Obs struct {
		Timestamp                 int64   // seconds
		WindLull                  float64 // m/s
		WindAvg                   float64 // m/s
		WindGust                  float64 // m/s
		WindDirection             int     // Degrees
		WindSampleInterval        int     // seconds
		StationPressure           float64 // MB
		AirTemperature            float64 // C
		RelativeHumidity          float64 // %
		Illuminance               int     // Lux
		UV                        float64 // Index
		SolarRadiation            int     // W/m*2
		PrecipitationAccumulation float64 // mm
		PrecipitationType         int     //
		StrikeAvgDistance         int     // km
		StrikeCount               int     // count
		Battery                   float64 // Voltags
		Interval                  int     // Minutes
	}
	var obs Obs

	for i := 0; i < 19; i++ {
		switch i {
		case 0:
			obs.Timestamp = int64(report.Obs[0][i])
		case 1:
			obs.WindLull = report.Obs[0][i]
		case 2:
			obs.WindAvg = report.Obs[0][i]
		case 3:
			obs.WindGust = report.Obs[0][i]
		case 4:
			obs.WindDirection = int(math.Round(report.Obs[0][i]))
		case 5:
			obs.WindSampleInterval = int(math.Round(report.Obs[0][i]))
		case 6:
			obs.StationPressure = report.Obs[0][i]
		case 7:
			obs.AirTemperature = report.Obs[0][i]
		case 8:
			obs.RelativeHumidity = report.Obs[0][i]
		case 9:
			obs.Illuminance = int(math.Round(report.Obs[0][i]))
		case 10:
			obs.UV = report.Obs[0][i]
		case 11:
			obs.SolarRadiation = int(math.Round(report.Obs[0][i]))
		case 12:
			obs.PrecipitationAccumulation = report.Obs[0][i]
		case 13:
			obs.PrecipitationType = int(math.Round(report.Obs[0][i]))
		case 14:
			obs.StrikeAvgDistance = int(math.Round(report.Obs[0][i]))
		case 15:
			obs.StrikeCount = int(math.Round(report.Obs[0][i]))
		case 16:
			obs.Battery = report.Obs[0][i]
		case 17:
			obs.Interval = int(math.Round(report.Obs[0][i]))
		}
	}
	if opts.Debug {
		log.Printf("OBS_ST %+v %+v", report, obs)
	}

	// Calculate Dew Point from RH and Temp
	dp, err := dewpoint.Calculate(obs.AirTemperature, obs.RelativeHumidity)
	if err != nil {
		log.Printf("dewpoint.Calculate(%f, %f): %v", obs.AirTemperature, obs.RelativeHumidity, err)
	}

	m.Timestamp = obs.Timestamp
	// Set fields and sort into alphabetical order to keep InfluxDB happy
	m.Fields = map[string]string{
		"battery":            fmt.Sprintf("%.2f", obs.Battery),
		"dew_point":          fmt.Sprintf("%.2f", dp),
		"illuminance":        fmt.Sprintf("%d", obs.Illuminance),
		"p":                  fmt.Sprintf("%.2f", obs.StationPressure),
		"precipitation":      fmt.Sprintf("%.2f", obs.PrecipitationAccumulation),
		"precipitation_type": fmt.Sprintf("%d", obs.PrecipitationType),
		"solar_radiation":    fmt.Sprintf("%d", obs.SolarRadiation),
		"strike_count":       fmt.Sprintf("%d", obs.StrikeCount),
		"strike_distance":    fmt.Sprintf("%d", obs.StrikeAvgDistance),
		"temp":               fmt.Sprintf("%.2f", obs.AirTemperature),
		"uv":                 fmt.Sprintf("%.2f", obs.UV),
		"wind_avg":           fmt.Sprintf("%.2f", obs.WindAvg),
		"wind_direction":     fmt.Sprintf("%d", obs.WindDirection),
		"wind_gust":          fmt.Sprintf("%.2f", obs.WindGust),
		"wind_lull":          fmt.Sprintf("%.2f", obs.WindLull),
	}
}

func tempest_rapid_wind(report Report, m *InfluxData) {
	type RapidWind struct {
		Timestamp     int64   // seconds
		WindSpeed     float64 // m/s
		WindDirection int     // degrees

	}
	var rapid_wind RapidWind

	for i := 0; i < 3; i++ {
		switch i {
		case 0:
			rapid_wind.Timestamp = int64(report.Ob[i])
		case 1:
			rapid_wind.WindSpeed = report.Ob[i]
		case 2:
			rapid_wind.WindDirection = int(math.Round(report.Ob[i]))
		}
	}
	if opts.Debug {
		log.Printf("RAPID_WIND %+v %+v", report, rapid_wind)
	}

	m.Timestamp = rapid_wind.Timestamp
	m.Fields = map[string]string{
		"rapid_wind_speed":     fmt.Sprintf("%.2f", rapid_wind.WindSpeed),
		"rapid_wind_direction": fmt.Sprintf("%d", rapid_wind.WindDirection),
	}
}

func tempest(addr *net.UDPAddr, b []byte, n int) (m *InfluxData, err error) {
	var report Report
	decoder := json.NewDecoder(bytes.NewReader(b[:n]))
	//		decoder.DisallowUnknownFields()
	err = decoder.Decode(&report)
	if err != nil {
		err = fmt.Errorf("ERROR Could not Unmarshal %d bytes from %v: %v: %v", n, addr, err, string(b[:n]))
		return
	}

	m = NewInfluxData()

	m.Bucket = opts.Influx_Bucket

	switch report.ReportType {
	case "obs_st":
		m.Name = "weather"
		tempest_obs_st(report, m)
		m.Tags["station"] = report.StationSerial
	case "rapid_wind":
		if !opts.Rapid_Wind {
			return
		}
		m.Name = "weather"
		tempest_rapid_wind(report, m)
		m.Tags["station"] = report.StationSerial
		if opts.Influx_Bucket_Rapid_Wind != "" {
			m.Bucket = opts.Influx_Bucket_Rapid_Wind
		}

	case "hub_status", "evt_precip", "evt_strike":
		return
	default:
		return
	}

	return
}
