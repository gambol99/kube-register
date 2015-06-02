package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/url"
)

var (
	apiEndpoint   string
	fleetEndpoint string
	metadata      string
	syncInterval  int
	healthzPort   string
	reverseLookup bool
	printVersion  bool
	nodelabels    bool
)

func init() {
	log.SetFlags(0)
	flag.StringVar(&apiEndpoint, "api-endpoint", "", "kubernetes API endpoint")
	flag.StringVar(&fleetEndpoint, "fleet-endpoint", "", "fleet endpoint")
	flag.StringVar(&metadata, "metadata", "k8s=kubelet", "comma-delimited key/value pairs")
	flag.StringVar(&healthzPort, "healthz-port", "10255", "the kubelet healthz port")
	flag.IntVar(&syncInterval, "sync-interval", 30, "sync interval")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.BoolVar(&reverseLookup, "reverse-lookup", false, "execute reverse lookup for registering hostnames instead of hosts' public IPs")
	flag.BoolVar(&nodelabels, "node-labels", false, "apply the fleet metedata as node labels on registration of minion")
}

func main() {
	flag.Parse()
	if printVersion {
		fmt.Printf("kube-register %s\n", Version)
		os.Exit(0)
	}
	m, err := parseMetadata(metadata)
	if err != nil {
		log.Println(err)
	}

	/* step: parse the url, otherwise fail */
	endpointURL, err := url.Parse(apiEndpoint)
	if err != nil {
		log.Printf("invalid api endpoint: %s, error: %s", apiEndpoint, err)
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		/* step: retrieve the registered machines from fleet */
		machines, err := getMachines(fleetEndpoint, healthzPort, m, reverseLookup)
		if err != nil {
			log.Println(err)
		}

		for _, machine := range machines {
			if err := register(endpointURL, machine); err != nil {
				log.Println(err)
			}
		}
		select {
		case c := <-signalChan:
			log.Println(fmt.Sprintf("captured %v exiting...", c))
			os.Exit(0)
		case <-time.After(time.Duration(syncInterval) * time.Second):
			// Continue syncing machines.
		}
	}
}
