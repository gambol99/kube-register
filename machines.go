package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/coreos/fleet/client"
)

type Machine struct {
	Name string
	Metadata map[string]string
}

func getMachines(endpoint, healthzPort string, metadata map[string][]string, reverseLookup bool) ([]Machine, error) {
	dialFunc := net.Dial
	machineList := make([]Machine, 0)
	/* step: parse the fleet endpoint url */
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "unix" {
		endpoint = "http://domain-sock/"
		dialFunc = func(network, addr string) (net.Conn, error) {
			return net.Dial("unix", u.Path)
		}
	}
	c := &http.Client{
		Transport: &http.Transport{
			Dial:              dialFunc,
			DisableKeepAlives: true,
		},
	}
	fleetClient, err := client.NewHTTPClient(c, endpoint)
	if err != nil {
		return nil, err
	}
	machines, err := fleetClient.Machines()
	if err != nil {
		return nil, err
	}
	for _, m := range machines {
		/* step: check if the machine has the metadata specified on the command line */
		if hasMetadata(m, metadata) && isHealthy(m.PublicIP, healthzPort) {
			machine := Machine{
				Name: m.PublicIP,
				Metadata: m.Metadata,
			}

			if reverseLookup {
				hostnames, err := net.LookupAddr(m.PublicIP)
				if err != nil {
					return nil, err
				}
				if len(hostnames) < 1 {
					return nil, errors.New("Could not get hostname for IP " + m.PublicIP)
				}
				hostname := hostnames[0]
				// remove trailing dot
				if hostname[len(hostname)-1] == '.' {
					hostname = hostname[:len(hostname)-1]
				}
				machine.Name = hostname
			}
			machineList = append(machineList, machine)
		}
	}
	return machineList, nil
}

func isHealthy(addr, healthzPort string) bool {
	url := fmt.Sprintf("http://%s:%s/healthz", addr, healthzPort)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("error health checking %s: %s", addr, err)
		return false
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
		return true
	}
	log.Printf("unhealthy machine: %s will not be registered", addr)
	return false
}
