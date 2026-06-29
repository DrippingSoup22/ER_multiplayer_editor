APP_NAME := er_pvp_mod
BIN_DIR  := bin
GUI_OUT  := $(BIN_DIR)/$(APP_NAME)_gui.exe
WIN_CC   := x86_64-w64-mingw32-gcc

.PHONY: all dirs gui clean

all: gui

dirs:
	mkdir -p $(BIN_DIR)

gui: dirs
	CC=$(WIN_CC) GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	go build -ldflags="-H windowsgui" -o $(GUI_OUT) ./cmd/gui

clean:
	rm -rf $(BIN_DIR)
