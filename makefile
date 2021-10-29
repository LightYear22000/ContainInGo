image = alpine
cmd = /bin/bash
build:
	go build -o cig
run:
	go build -o cig && sudo ./cig run $(image) $(cmd)