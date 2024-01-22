package main

import (
	"fmt"
	"os"
)

type CliInvalidUsageError struct{}

func (e CliInvalidUsageError) Error() string {
	return fmt.Sprintf("Invalid usage, type %s -h to get help", os.Args[0])
}

type CliService struct {
	tokenSvc *TokenService
	args     []string
}

func (svc *CliService) ExecuteCliCommands() (string, error) {
	if len(svc.args) < 3 || len(svc.args) > 4 {
		return "", CliInvalidUsageError{}
	}

	if svc.args[1] != "token" {
		return "", CliInvalidUsageError{}
	}

	var (
		isBoard bool
		sub     string
	)

	if len(svc.args) == 4 {
		if svc.args[3] != "-b" {
			return "", CliInvalidUsageError{}
		}

		isBoard = true
	} else {
		isBoard = false
	}

	sub = svc.args[2]

	return svc.tokenSvc.New(&TokenSubject{
		sub,
		isBoard,
	})
}
