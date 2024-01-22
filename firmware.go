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

func (svc *FirmwareService) AddFirmware(info *FirmwareInfo, bytes []byte) (*FirmwareInfo, error) {
	// TODO: AES encryption.
	h := sha256.New()
    h.Write([]byte(bytes))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	if info.Sha256 == "" {
		info.Sha256 = hash
	} else if hash != info.Sha256 {
		return nil, SHA256DiffersError{given: info.Sha256, computed: hash}
	}

	addedInfo, err := svc.db.AddFirmwareInfo(info)
	if err != nil {
		return nil, err
	}

	if err := svc.bins.AddFirmwareBinary(addedInfo.Id, bytes); err != nil {
		// TODO: remove record from db.
		return nil, err
	}

	return addedInfo, nil
}

func (serv *FirmwareService) GetNewestFirmware(repo string, tags []string) (*FirmwareInfo, error) {
	return serv.db.GetNewestFirmwareInfo(repo, tags)
}

// func (serv *FirmwareService) GetFirmwareBinary(firmware_id int64) ([]byte, error) {
// 	// TODO: AES encryption.
// 	return serv.bins.GetFirmwareBinary(firmware_id)
// }

func (serv *FirmwareService) GetFirmwaresInfo() ([]FirmwareInfo, error) {
	return serv.db.GetAllFirmwares()
}
