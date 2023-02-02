#!/bin/sh

nitriding \
	-acme \
	-fqdn "nitro.nymity.ch" \
	-appurl "https://github.com/brave-experiments/vct" \
	-appwebsrv "http://127.0.0.1:8080" \
	-extport 8443 \
	-intport 8081 &
echo "[sh] Started nitriding as reverse proxy."

sleep 1

vct
echo "[sh] Started vct."
