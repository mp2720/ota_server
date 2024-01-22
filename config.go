package main

import "gopkg.in/ini.v1"

type Config struct {
	storagePath           string
	host                  string
	port                  string
	jwtSigningKey         string
	jwtIssuer             string
	firmwareEncryptionKey string
}

func LoadConfig() (*Config, error) {
	iniFile, err := ini.Load("config.ini")
	if err != nil {
		return nil, err
	}

	return &Config{
		storagePath:           iniFile.Section("").Key("storagePath").String(),
		host:                  iniFile.Section("").Key("host").String(),
		port:                  iniFile.Section("").Key("port").String(),
		jwtSigningKey:         iniFile.Section("jwt").Key("signingKey").String(),
		jwtIssuer:             iniFile.Section("jwt").Key("issuer").String(),
		firmwareEncryptionKey: iniFile.Section("firmware").Key("encryptionKey").String(),
	}, nil
}
