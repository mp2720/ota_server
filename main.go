package main

import (
	"os"
)

const STORAGE_PATH = "./storage"

func main() {
	// Create if not exists
	err := os.MkdirAll(STORAGE_PATH, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
