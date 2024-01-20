package main

import (
	"os"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	// Create if not exists
	err = os.MkdirAll(cfg.storagePath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
