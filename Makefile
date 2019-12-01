build:
	go generate
	go build .
	strip members
