package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"sort"
	"strings"

	"github.com/de-wax/go-pkg/dewpoint"
)

var PrecipitationType []string = []string{"none","rain","hail","rain+hail"}

type Obs struct {
	Timestamp int64				// seconds
	WindLull float64			// m/s
	WindAvg float64				// m/s
	WindGust float64			// m/s
	WindDirection int			// Degrees
	WindSampleInterval int			// seconds
	StationPressure float64			// MB
	AirTemperature float64			// C
	RelativeHumidity float64		// %
	Illuminance int				// Lux
	UV float64				// Index
	SolarRadiation int			// W/m*2
	PrecipitationAccumulation float64	// mm
	PrecipitationType int			//
	StrikeAvgDistance int			// km
	StrikeCount int 			// count
	Battery float64				// Voltags
	Interval int				// Minutes
}

type Report struct {
	StationSerial string	`json:"serial_number,omitempty"`
	ReportType string	`json:"type"`
	HubSerial string	`json:"hub_sn,omitempty"`
	Obs [1][] float64	`json:"obs,omitempty"`
	Ob [3]float64		`json:"ob,omitempty"`
	//	Firmware_revision string `json:"firmware_revision,omitempty,string"`
	Uptime int		`json:"uptime,omitempty"`
	Timestamp int		`json:"timestamp,omitempty"`
	ResetFlags string	`json:"reset_flags,omitempty"`
	Seq int			`json:"seq,omitempty"`
	Fs []float64		`json:"fs,omitempty"`
	Radio_Stats []float64	`json:"radio_stats,omitempty"`
	Mqtt_Stats []float64	`json:"mqtt_stats,omitempty"`
	Voltage float64		`json:"voltage,omitempty"`
	RSSI float64		`json:"rssi,omitempty"`
	HubRSSI float64		`json:"hub_rssi,omitempty"`
	SensorStatus int	`json:"sensor_status,omitempty"`
	Debug int		`json:"debug,omitempty"`
}

func tempest(logger *log.Logger, addr *net.UDPAddr, b []byte, n int) string {
	var report Report
	decoder := json.NewDecoder(bytes.NewReader(b[:n]))
		//		decoder.DisallowUnknownFields()
	err := decoder.Decode(&report)
	if err != nil {
		logger.Printf("Could not Unmarshal %d bytes from %v: %v: %v", n, addr, err, string(b[:n]))
		return ""
	}

	if report.ReportType != "obs_st" {
		return ""
	}

	var obs Obs
	for i:=0; i<19; i++ {
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
	logger.Printf("REPORT %+v %+v", report, obs)

	// Calculate Dew Point from RH and Temp
	dp, err := dewpoint.Calculate(obs.AirTemperature, obs.RelativeHumidity)
	if err != nil {
		logger.Printf("dewpoint.Calculate(%f, %f): %v", obs.AirTemperature, obs.RelativeHumidity, err)
	}

	// Set fields and sort into alphabetical order to keep InfluxDB happy
	fields := map[string]string{
		"battery":		fmt.Sprintf("%.2f", obs.Battery),
		"dew_point":		fmt.Sprintf("%.2f", dp),
		"illuminance":		fmt.Sprintf("%d", obs.Illuminance),
		"p":			fmt.Sprintf("%.2f", obs.StationPressure),
		"precipitation":	fmt.Sprintf("%.2f", obs.PrecipitationAccumulation),
		"precipitation_type":	fmt.Sprintf("%d", obs.PrecipitationType),
		"solar_radiation":	fmt.Sprintf("%d", obs.SolarRadiation),
		"strike_count":         fmt.Sprintf("%d", obs.StrikeCount),
		"strike_distance":      fmt.Sprintf("%d", obs.StrikeAvgDistance),
		"temp":			fmt.Sprintf("%.2f", obs.AirTemperature),
		"uv":			fmt.Sprintf("%.2f", obs.UV),
		"wind_avg":		fmt.Sprintf("%.2f", obs.WindAvg),
		"wind_direction":	fmt.Sprintf("%d", obs.WindDirection),
		"wind_gust":		fmt.Sprintf("%.2f", obs.WindGust),
		"wind_lull":		fmt.Sprintf("%.2f", obs.WindLull),
	}
	field_list := make([]string, 0, len(fields))
	for k := range fields {
		field_list = append(field_list, fmt.Sprintf("%s=%s", k, fields[k]))
	}
	sort.Strings(field_list)

	line := fmt.Sprintf("weather,station=%s %s %v\n",
		report.StationSerial,
		strings.Join(field_list, ","),
		obs.Timestamp)
	
	return line
}
