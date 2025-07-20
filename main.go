package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Services []string `yaml:"services"`
}

type Result struct {
	Service string
	Status  string // UP or DOWN
	Method  string // http, https, tcp, ping, or error description
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func checkHTTP(service string) (bool, string) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(service)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	return true, "http"
}

func checkTCP(service string) (bool, string) {
	conn, err := net.DialTimeout("tcp", service, 2*time.Second)
	if err != nil {
		return false, err.Error()
	}
	_ = conn.Close()
	return true, "tcp"
}

func checkPing(host string) (bool, string) {
	out, err := exec.Command("ping", "-c", "1", "-W", "2", host).Output()
	if err != nil {
		return false, err.Error()
	}
	if strings.Contains(string(out), "1 received") || strings.Contains(string(out), "1 packets received") {
		return true, "ping"
	}
	return false, "ping timeout"
}

func getHostPart(address string) string {
	if strings.HasPrefix(address, "http") {
		addr := strings.Split(address, "://")[1]
		addr = strings.Split(addr, "/")[0]
		return strings.Split(addr, ":")[0]
	}
	return strings.Split(address, ":")[0]
}

func checkService(service string) Result {
	if strings.HasPrefix(service, "http://") || strings.HasPrefix(service, "https://") {
		ok, reason := checkHTTP(service)
		if ok {
			return Result{service, "UP", reason}
		}
		// fallback to TCP
		hostPort := strings.Split(service, "://")[1]
		ok, reason = checkTCP(hostPort)
		if ok {
			return Result{service, "UP", reason}
		}
		// fallback to ping
		host := getHostPart(service)
		ok, reason = checkPing(host)
		if ok {
			return Result{service, "UP", reason}
		}
		return Result{service, "DOWN", reason}
	} else {
		// assume host:port
		ok, reason := checkTCP(service)
		if ok {
			return Result{service, "UP", reason}
		}
		host := getHostPart(service)
		ok, reason = checkPing(host)
		if ok {
			return Result{service, "UP", reason}
		}
		return Result{service, "DOWN", reason}
	}
}

func main() {
	configFlag := flag.String("c", "", "Path to config file (optional). Defaults to ~/.health-checker/config.yaml or ./config.yaml")
	flag.Parse()

	// Resolve config path
	var configPaths []string
	if *configFlag != "" {
		configPaths = append(configPaths, *configFlag)
	} else {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPaths = append(configPaths, homeDir+"/.health-checker/config.yaml")
		}
		configPaths = append(configPaths, "config.yaml")
	}

	var config *Config
	var err error
	for _, path := range configPaths {
		config, err = loadConfig(path)
		if err == nil {
			break
		}
	}

	if config == nil {
		fmt.Fprintf(os.Stderr, "Could not find valid config file. Tried: %v\n", configPaths)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	results := make(chan Result, len(config.Services))

	for _, service := range config.Services {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			results <- checkService(s)
		}(service)
	}

	wg.Wait()
	close(results)

	fmt.Println("Service Status Report:")
	for r := range results {
		fmt.Printf("%-30s -> %-4s (%s)\n", r.Service, r.Status, r.Method)
	}
}

