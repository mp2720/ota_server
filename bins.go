package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type BinariesService struct {
	cfg *Config
}

func (svc *BinariesService) getFirmwareBinaryPath(firmware_id int64) string {
	return filepath.Join(svc.cfg.storagePath, fmt.Sprintf("%d.bin", firmware_id))
}

func (svc *BinariesService) AddFirmwareBinary(firmware_id int64, bytes []byte) error {
	f, err := os.Create(svc.getFirmwareBinaryPath(firmware_id))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	return err
}

func (svc *BinariesService) GetFirmwareBinary(firmware_id int64) ([]byte, error) {
	return os.ReadFile(svc.getFirmwareBinaryPath(firmware_id))
}

