package main

import (
	"log"
	"net"
	"net/url"
	"net/http"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

var opts struct {
	Source string
	Target string
	Token string
	Bucket string
	Buffer int
	Verbose bool
	Debug bool
}

func packet(logger *log.Logger, url *url.URL, addr *net.UDPAddr, b []byte, n int) {
	line := tempest(logger, addr, b, n)
	if line == "" {
		return
	}

	if opts.Verbose {
		logger.Printf("POST %s", line)
	}

	request, err := http.NewRequest("POST", url.String(), strings.NewReader(line))
	if err != nil {
		logger.Printf("NewRequest: %v", err)
		return
	}
	request.Header.Set("Authorization", "Token " + opts.Token)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		logger.Printf("Posting to %s: %v", opts.Target, err)
		return
	}
	if resp.StatusCode >= 400 {
		logger.Printf("POST: %s", resp.Status)
	}
	resp.Body.Close()
}

func parse() {
	flag.StringVar(&opts.Source, "source", ":50222", "Source port to listen on")
	flag.StringVar(&opts.Target, "target", "https://localhost:50222/api/v2/write", "URL to receive influx metrics")
	flag.StringVar(&opts.Token, "token", "", "Authentication token")
	flag.StringVar(&opts.Bucket, "bucket", "", "InfluxDB bucket name")
	flag.IntVar(&opts.Buffer, "buffer", 10240, "Max buffer size for the socket io")
	flag.BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose logging")
	flag.BoolVarP(&opts.Debug, "debug", "d", false, "Debug logging")

	flag.Parse()
	if opts.Debug {
		opts.Verbose = opts.Debug
	}
}

func main() {
	logger := log.New(os.Stdout, "tempest_influx: ", log.LstdFlags)

	parse()

	sourceAddr, err := net.ResolveUDPAddr("udp", opts.Source)
	if err != nil {
		logger.Fatalf("Could not resolve source address: %s: %s", opts.Source, err)
	}

	sourceConn, err := net.ListenUDP("udp", sourceAddr)
	if err != nil {
		logger.Fatalf("Could not listen on address: %s: %s", opts.Source, err)
		return
	}

	defer sourceConn.Close()

	url, err := url.Parse(opts.Target)
	query := url.Query()
	query.Set("precision", "s")
	if opts.Bucket != "" {
		query.Set("bucket", opts.Bucket)
	}
	url.RawQuery = query.Encode()

	logger.Printf(">> Starting tempest_influx, Verbose %v Debug %v Source at %v, Target at %v",
		opts.Verbose,
		opts.Debug,
		opts.Source,
		url.String())

	for {
		b := make([]byte, opts.Buffer)
		n, addr, err := sourceConn.ReadFromUDP(b)
		if err != nil {
			logger.Printf("Could not receive a packet from %s: %s", addr, err)
			continue
		}

		if opts.Debug {
			logger.Printf("\nRECV %v %d bytes: %s", addr, n, string(b[:n]))
		}

		go packet(logger, url, addr, b, n)
	}
}
