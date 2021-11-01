build:
	go build ./cmd/edead
	(cd frontend; TERM=xterm-256color ./build-fe.sh)

clean:
	go clean
	rm edead
