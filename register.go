package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Node struct {
	Kind       string   `json:"kind,omitempty"`
	APIVersion string   `json:"apiVersion,omitempty"`
	Metadata   Metadata `json:"metadata,omitempty"`
	Spec       Spec     `json:"spec,omitempty"`
}

type Metadata struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Spec struct {
	ExternalID string `json:"externalID,omitempty"`
}

type NodeResp struct {
	Reason string `json:"reason,omitempty"`
}

func register(endpoint *url.URL, machine Machine) error {
	var n Node
	n.Kind = "Node"
	n.APIVersion = apiVersion
	n.Metadata.Name = machine.Name
	if nodelabels {
		n.Metadata.Labels = machine.Metadata
	}
	n.Spec.ExternalID = machine.Name

	data, err := json.Marshal(&n)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s://%s/api/%s/nodes", endpoint.Scheme, endpoint.Host, apiVersion)

	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 202 || res.StatusCode == 200 || res.StatusCode == 201 {
		log.Printf("registered machine: %s\n", machine.Name)
		return nil
	}

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 409 {
		return fmt.Errorf("error registering: %s %d %s", machine.Name, res.StatusCode, string(data))
	}

	nr := &NodeResp{}
	if err := json.Unmarshal([]byte(data), &nr); err != nil {
		return err
	}

	if res.StatusCode == 409 && nr.Reason == "AlreadyExists" {
		return nil
	}

	return fmt.Errorf("error registering: %s %s", machine.Name, nr.Reason)
}
