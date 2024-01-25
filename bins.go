package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type BinariesService struct {
	cfg *Config
}

func (svc *BinariesService) GetFirmwareBinaryPath(uuid string) string {
	return filepath.Join(svc.cfg.storagePath, fmt.Sprintf("%s.bin", uuid))
}

func (svc *BinariesService) AddFirmwareBinary(uuid string, bytes []byte) error {
	f, err := os.Create(svc.GetFirmwareBinaryPath(uuid))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	return err
}
