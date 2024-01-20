package main

import (
	"crypto/sha256"
	"fmt"
)

type FirmwareService struct {
	db DB
}

type SHA256DiffersError struct {
	given    string
	computed string
}

func (e SHA256DiffersError) Error() string {
	return fmt.Sprintf("SHA256 %s (given) != %s (computed)", e.given, e.computed)
}

func (serv *FirmwareService) AddFirmware(info *FirmwareInfo, bytes []byte) error {
	h := sha256.New()
	hash := fmt.Sprintf("%x", h.Sum(bytes))
	if info.sha256 == "" {
		info.sha256 = hash
	} else if hash != info.sha256 {
		return SHA256DiffersError{given: info.sha256, computed: hash}
	}

	id, err := serv.db.AddFirmwareInfo(info)
	if err != nil {
		return err
	}

	return AddFirmwareBinary(id, bytes)
}
