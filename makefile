image = alpine
cmd = /bin/sh
build:
	go build -o cig
run:
	go build -o cig && sudo ./cig run $(image) $(cmd)
clean:
	sudo rm -rf /var/lib/cig