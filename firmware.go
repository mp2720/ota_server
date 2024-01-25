package main

import (
	"crypto/md5"
	"fmt"
	guuid "github.com/google/uuid"
)

type FirmwareService struct {
	db   *DB
	bins *BinariesService
}

type Md5DiffersError struct {
	given    string
	computed string
}

func (e *Md5DiffersError) Error() string {
	return fmt.Sprintf("MD5 %s (given) != %s (computed)", e.given, e.computed)
}

type FirmwareNotFoundError struct{}

func (e *FirmwareNotFoundError) Error() string {
	return "firmware not found"
}

type FirmwareFileAlreadyUploaded struct{}

func (e *FirmwareFileAlreadyUploaded) Error() string {
	return "firmware file already uploaded"
}

func (svc *FirmwareService) CreateFirmware(info *FirmwareInfo) (*FirmwareInfo, error) {
	info.Size = 0
	info.Uuid = guuid.New().String()
	return svc.db.AddFirmwareInfo(info)
}

func (svc *FirmwareService) AddFirmwareFile(uuid string, bytes []byte) error {
	info, err := svc.db.GetFirmareInfoByUuid(uuid)
	if err != nil {
		return err
	}
	if info == nil {
		return &FirmwareNotFoundError{}
	}

	if info.hasBin() {
		return &FirmwareFileAlreadyUploaded{}
	}

	h := md5.New()
	h.Write([]byte(bytes))
	info.Md5 = fmt.Sprintf("%x", h.Sum(nil))
	info.Size = len(bytes)

	if err := svc.bins.AddFirmwareBinary(uuid, bytes); err != nil {
		return err
	}

	return svc.db.UpdateFirmwareFileInfo(info)
}

func (serv *FirmwareService) GetLatestFirmware(repo string, board string) (*FirmwareInfo, error) {
	return serv.db.GetLatestFirmwareInfo(repo, board)
}

func (serv *FirmwareService) GetFirmwareBinaryPath(uuid string) (string, error) {
	fi, err := serv.db.GetFirmareInfoByUuid(uuid)
	if err != nil || fi == nil || !fi.hasBin() {
		return "", err
	}

	return serv.bins.GetFirmwareBinaryPath(uuid), nil
}

func (serv *FirmwareService) GetAllFirmwaresInfo() ([]FirmwareInfo, error) {
	return serv.db.GetAllFirmwaresInfo()
}
