APP_NAME := vk
MAIN := ./cmd/$(APP_NAME)/main.go

run:
	go run $(MAIN)