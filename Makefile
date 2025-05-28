build: test
	go generate
	go build -ldflags "-w -s" -trimpath .

test:
	go test -coverprofile members.coverage ./...
