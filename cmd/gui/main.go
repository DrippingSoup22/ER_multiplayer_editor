package main

import (
	"log"

	"er_pvp_mod/ui/windows"
)

func main() {
	if err := windows.RunGUI(); err != nil {
		log.Fatal(err)
	}
}
