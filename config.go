package main

import "gopkg.in/ini.v1"

type Config struct {
	storagePath           string
	jwtSigningKey         string
	firmwareEncryptionKey string
}

func LoadConfig() (*Config, error) {
	iniFile, err := ini.Load("config.ini")
	if err != nil {
		return nil, err
	}

	return &Config{
		storagePath:           iniFile.Section("").Key("storagePath").String(),
		jwtSigningKey:         iniFile.Section("jwt").Key("signingKey").String(),
		firmwareEncryptionKey: iniFile.Section("firmware").Key("encryptionKey").String(),
	}, nil
}
