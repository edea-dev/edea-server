build:
	go build ./cmd/edead
	(cd frontend; TERM=xterm-256color ./build-fe.sh)

live-backend:
	go build ./cmd/edead

live-frontend:
	(cd frontend; TERM=xterm-256color ./build-fe.sh)

test:
	(cd frontend; npx playwright test)

clean:
	go clean
	rm edead
