package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func getFirmwareBinaryPath(firmware_id int64) string {
	return filepath.Join(STORAGE_PATH, fmt.Sprintf("%d.bin", firmware_id))
}

func AddFirmwareBinary(firmware_id int64, bytes []byte) error {
	f, err := os.Create(getFirmwareBinaryPath(firmware_id))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	return err
}

func GetFirmwareBinary(firmware_id int64) ([]byte, error) {
	return os.ReadFile(getFirmwareBinaryPath(firmware_id))
}
