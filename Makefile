.PHONY: all

name = vct
godeps = go.mod *.go
docker-image = $(name):latest
enclave-file = $(name).eif
gvproxy_port = 443
nitriding_port = 8443

all: run-enclave

.PHONY: gvproxy
gvproxy:
	sudo killall -q gvproxy | true # Ignore non-zero exit status.
	sudo gvproxy \
		-listen vsock://:1024 \
		-listen unix:///tmp/network.sock \
		-mtu 65000 &
	sleep 1
	sudo curl \
		--unix-socket /tmp/network.sock \
		-X POST \
		-d '{"local":":$(gvproxy_port)","remote":"192.168.127.2:$(nitriding_port)"}' \
		http:/unix/services/forwarder/expose

.PHONY: nitriding
nitriding:
	make -C nitriding/cmd/

$(name): $(godeps)
	make -C example/

.PHONY: docker
docker: nitriding $(name) Dockerfile
	docker build -t $(name):latest .

$(enclave-file): docker
	nitro-cli build-enclave \
		--docker-uri $(docker-image) \
		--output-file $(enclave-file)

.PHONY: kill-enclave
kill-enclave:
	$(eval ENCLAVE_ID=$(shell nitro-cli describe-enclaves | jq -r '.[0].EnclaveID'))
	@if [ "$(ENCLAVE_ID)" != "null" ]; then nitro-cli terminate-enclave --enclave-id $(ENCLAVE_ID); fi

.PHONY: run-enclave
run-enclave: $(enclave-file) kill-enclave gvproxy
	nitro-cli run-enclave \
		--cpu-count 2 \
		--memory 4000 \
		--enclave-cid 4 \
		--eif-path $(enclave-file) \
		--debug-mode \
		--attach-console

.PHONY: clean
clean:
	make -C nitriding/cmd/ clean
	make -C example/ clean
	rm -f $(enclave-file)
