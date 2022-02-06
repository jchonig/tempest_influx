package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var opts *Config

func packet(url *url.URL, addr *net.UDPAddr, b []byte, n int) {
	line := tempest(addr, b, n)
	if line == "" {
		return
	}

	if opts.Verbose {
		log.Printf("POST %s", line)
	}

	request, err := http.NewRequest("POST", url.String(), strings.NewReader(line))
	if err != nil {
		log.Printf("NewRequest: %v", err)
		return
	}
	request.Header.Set("Authorization", "Token "+opts.Influx_Token)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Printf("Posting to %s: %v", opts.Influx_URL, err)
		return
	}
	if resp.StatusCode >= 400 {
		log.Printf("POST: %s", resp.Status)
	}
	resp.Body.Close()
}

func main() {
	log.SetPrefix("tempest_influx: ")

	opts = LoadConfig("/config", "tempest_influx")
	if opts.Debug {
		spew.Dump(opts)
	}

	sourceAddr, err := net.ResolveUDPAddr("udp", opts.Listen_Address)
	if err != nil {
		log.Fatalf("Could not resolve source address: %s: %s", opts.Listen_Address, err)
	}

	sourceConn, err := net.ListenUDP("udp", sourceAddr)
	if err != nil {
		log.Fatalf("Could not listen on address: %s: %s", opts.Listen_Address, err)
		return
	}

	defer sourceConn.Close()

	url, err := url.Parse(opts.Influx_URL)
	query := url.Query()
	query.Set("precision", "s")
	if opts.Influx_Bucket != "" {
		query.Set("bucket", opts.Influx_Bucket)
	}
	url.RawQuery = query.Encode()

	log.Printf(">> Starting tempest_influx, Verbose %v Debug %v Listen_Address %v, Target %v",
		opts.Verbose,
		opts.Debug,
		opts.Listen_Address,
		url.String())

	for {
		b := make([]byte, opts.Buffer)
		n, addr, err := sourceConn.ReadFromUDP(b)
		if err != nil {
			log.Printf("Could not receive a packet from %s: %s", addr, err)
			continue
		}

		if opts.Debug {
			log.Printf("\nRECV %v %d bytes: %s", addr, n, string(b[:n]))
		}

		go packet(url, addr, b, n)
	}
}
