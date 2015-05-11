# kube-register

Register Kubernetes Kubelet machines with the Kubernetes API server using Fleet data.

## Usage

```
kube-register -h
Usage of ./kube-register:
  -api-endpoint="": kubernetes API endpoint
  -fleet-endpoint="": fleet endpoint
  -healthz-port="10255": the kubelet healthz port
  -metadata="k8s=kubelet": comma-delimited key/value pairs
  -reverse-lookup=false: execute reverse lookup for registering hostnames instead of hosts' public IPs
  -sync-interval=30: sync interval
  -version=false: print version and exit
```

### Example

```
kube-register -metadata="kubelet=true" -fleet-endpoint="http://127.0.0.1:4002" -apiserver-endpoint="http://127.0.0.1:8080"
```

By default kube-register registers new machines with their public IP addresses, found in fleet. By setting `-reverse-lookup=true` kube-register will do a reverse DNS lookup to get the machine's hostname and registers the machine with its hostname instead with its IP address. Make sure the machine's IP resolves to the same hostname that is printed by `hostname -f` on the machine itself.

### Enable the fleet endpoint

The kube-register services requires access to a fleet-endpoint. For example on a CoreOS system, enable the fleet end-point with the following systemd socket file:

```
[Socket]
ListenStream=/var/run/fleet.sock
ListenStream=8000
Service=fleet.service

[Install]
WantedBy=sockets.target
```

## Building

```
mkdir -p "${GOPATH}/src/github.com/kelseyhightower"
cd "${GOPATH}/src/github.com/kelseyhightower"
git clone https://github.com/kelseyhightower/kube-register.git
cd kube-register
godep go build .
```

Slightly smaller binary.

```
CGO_ENABLED=0 GOOS=linux godep go build -a -tags netgo -ldflags '-w' .
```

To build the binary inside a docker container.

```
docker run -v $SRC:/opt/kube-register -i -t google/golang /bin/bash -c "cd /opt/kube-register && go get github.com/tools/godep && godep go build ."
```

Where $SRC is the absolute path to the kube-register directory.
