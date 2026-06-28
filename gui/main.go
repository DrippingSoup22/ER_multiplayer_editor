package main

import (
	"log"

	"er_pvp_mod/app"
)

func main() {
	if err := app.RunGUI(); err != nil {
		log.Fatal(err)
	}
}
