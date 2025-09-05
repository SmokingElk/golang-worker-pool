ENTRY_PATH = ./cmd/app/main.go
OUT_PATH = ./worker-pull-app.exe

run:
	go run $(ENTRY_PATH)

build:
	go build $(ENTRY_PATH) -o $(OUT_PATH)