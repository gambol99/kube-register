.PHONY: build

build:
	mkdir -p ./bin 
	godep go build -o bin/kube-register

clean:
	rm -rf ./bin