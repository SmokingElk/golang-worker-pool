ENTRY_PATH = ./cmd/app/main.go
OUT_PATH = ./worker-pool-app.exe
TEST_PATH = ./internal/worker_pool/

run:
	go run $(ENTRY_PATH)

build:
	go build $(ENTRY_PATH) -o $(OUT_PATH)

test: 
	go test -v -count=1 $(TEST_PATH)

test10: 
	go test -v -count=10 $(TEST_PATH)

cover:
	go test -coverprofile=cover.out -count=1 $(TEST_PATH)
	go tool cover -html=cover.out
	DEL cover.out