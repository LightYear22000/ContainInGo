run:
	go build -o cig && sudo ./cig run ubuntu:latest /bin/bash

build:
	go build -o cig