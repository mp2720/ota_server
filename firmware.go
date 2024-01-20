package main

import (
	"crypto/sha256"
	"fmt"
)

type FirmwareService struct {
	db   *DB
	bins *BinariesService
}

type SHA256DiffersError struct {
	given    string
	computed string
}

func (e SHA256DiffersError) Error() string {
	return fmt.Sprintf("SHA256 %s (given) != %s (computed)", e.given, e.computed)
}

func (svc *FirmwareService) AddFirmware(info *FirmwareInfo, bytes []byte) error {
	// TODO: AES encryption.
	h := sha256.New()
	hash := fmt.Sprintf("%x", h.Sum(bytes))
	if info.sha256 == "" {
		info.sha256 = hash
	} else if hash != info.sha256 {
		return SHA256DiffersError{given: info.sha256, computed: hash}
	}

	id, err := svc.db.AddFirmwareInfo(info)
	if err != nil {
		return err
	}

	return svc.bins.AddFirmwareBinary(id, bytes)
}

func (serv *FirmwareService) GetNewestFirmware(repo string, tags []string) (*FirmwareInfo, error) {
	return serv.db.GetNewestFirmwareInfo(repo, tags)
}

func (serv *FirmwareService) GetFirmwareBinary(firmware_id int64) ([]byte, error) {
	// TODO: AES encryption.
	return serv.bins.GetFirmwareBinary(firmware_id)
}
