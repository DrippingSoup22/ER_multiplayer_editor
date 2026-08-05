APP_NAME := er_pvp_mod
BIN_DIR  := bin
GUI_OUT  := $(BIN_DIR)/$(APP_NAME)_gui.exe

.PHONY: all dirs gui release clean

all: gui

dirs:
	mkdir -p $(BIN_DIR)

gui: dirs
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
	go build -trimpath -ldflags="-H windowsgui -s -w" -o $(GUI_OUT) ./cmd/gui

release:
	bash scripts/build-windows.sh

clean:
	rm -rf $(BIN_DIR) dist
