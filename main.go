package main

import (
	"fmt"
	"os"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	// Create if not exists
	err = os.MkdirAll(cfg.storagePath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	tokenSvc := TokenService{cfg}

	if len(os.Args) == 1 {
		db, err := NewDB(cfg)
		if err != nil {
			panic(err)
		}
		binSvc := BinariesService{cfg}
		firmwareSvc := FirmwareService{
			db,
			&binSvc,
		}
		api := Api{
			&firmwareSvc,
			&binSvc,
            &tokenSvc,
			cfg,
		}
		if err := api.StartServer(); err != nil {
			panic(err)
		}
	} else if len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Printf("%s - launch HTTP server\n", os.Args[0])
		fmt.Printf("%s token <subject-name> [-b] - generate JWT for subject\n", os.Args[0])
		fmt.Printf("\t-b - if subject is board\n")
		os.Exit(0)
	} else {
        cliSvc := CliService{
            &tokenSvc,
            os.Args,
        }
        result, err := cliSvc.ExecuteCliCommands()
        if err != nil {
            panic(err)
        }

        fmt.Println(result)
	}
}
