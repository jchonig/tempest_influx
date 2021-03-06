package main

import (
	"github.com/spf13/viper"
	"log"

	flag "github.com/spf13/pflag"
)

type Config struct {
	Config_Dir               string `mapstructure:"CONFIG_DIR"`
	Listen_Address           string `mapstructure:"LISTEN_ADDRESS"`
	Influx_URL               string `mapstructure:"INFLUX_URL"`
	Influx_Token             string `mapstructure:"INFLUX_TOKEN"`
	Influx_Bucket            string `mapstructure:"INFLUX_BUCKET"`
	Influx_Bucket_Rapid_Wind string `mapstructure:"INFLUX_BUCKET_RAPID_WIND"`
	Buffer                   int
	Verbose                  bool
	Debug                    bool
	Noop                     bool
	Rapid_Wind               bool `mapstructure:"RAPID_WIND"`
}

func LoadConfig(path string, name string) (config *Config) {
	config_file := name + ".yml"

	viper.SetDefault("Listen_Address", ":50222")
	viper.SetDefault("Influx_URL", "https://localhost:8086/api/v2/write")
	viper.SetDefault("Buffer", 10240)

	flag.String("listen_address", "", "Address to listen for UDP Broadcasts")
	flag.String("influx_url", "", "URL to receive influx metrics")
	flag.String("influx_token", "", "Authentication token for Influx")
	flag.String("influx_bucket", "", "InfluxDB bucket name")
	flag.String("influx_bucket_rapid_wind", "", "InfluxDB bucket name for rapid wind reports")
	flag.Int("buffer", 0, "Max buffer size for the socket io")
	flag.BoolP("verbose", "v", false, "Verbose logging")
	flag.BoolP("debug", "d", false, "Debug logging")
	flag.BoolP("noop", "n", false, "Don't post to influx")
	flag.Bool("rapid_wind", false, "Send rapid wind reports")

	viper.AddConfigPath(path)

	viper.SetConfigName(config_file)
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix(name)
	viper.AutomaticEnv()

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)
	if viper.GetBool("debug") {
		viper.Set("verbose", true)
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			log.Fatalf("%v", err)
		}
	}
	err = viper.Unmarshal(&config)

	return config
}
