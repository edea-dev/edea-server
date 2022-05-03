deps:
	(cd frontend; ./install-dependencies.sh)

build:
	(cd frontend; TERM=xterm-256color ./build-fe.sh)
	go build ./cmd/edea-server

live-backend:
	go build ./cmd/edea-server

live-frontend:
	(cd frontend; TERM=xterm-256color ./build-fe.sh)

test:
	(cd frontend; npx playwright test)

clean:
	go clean
	rm edead
