#!/bin/bash

export GOPATH=/home/wolf/Maximuma/lib:/home/wolf/Maximuma/Streamer
#CONF=/home/wolf/Maximuma/Streamer/config/urls.json
CONF=/opt/streamer/urls.json.tiny
go run src/Streamer.go -sources=${CONF} -networks=/opt/streamer/networks.json -enable-web-controls
