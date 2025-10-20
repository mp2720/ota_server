# OTA Server
A firmware version control server for OTA (Over The Air) updates.

Build:
```bash
go build
```

Configuration must be located in the `config.ini` file; an example is provided in `config.ini.example`.

Start the server:
```bash
./ota_server
```

The server stores data in the folder specified in the config.
This folder contains the sqlite database with information about each version and the firmware files.

## API
API documentation can be found at `/swagger/index.html`.

Developers can upload new firmware versions.
Additional information required for each firmware includes:
* build time
* git commit hash (optional)
* repository name
* list of board names the firmware can be uploaded to
* description (can be empty)

Boards can request the latest firmware version, providing the repository name.

## Security
To use the HTTP API, you need to generate JWT tokens.
They contain the subject name for whom the token is issued and its type (developer/board).
The secret key must be specified in the configuration.

Token for a developer (grants access to: uploading firmware to the server, viewing all firmware):
```
./ota_server token %USERNAME%
```

Token for a board (grants access only to getting the latest firmware version):
```
./ota_server token %BOARDNAME% -b
```

## TLS
The config must specify the paths to the .pem and .key files (in the example, these are `./tls/ota_server.key|pem`)

You can generate them like this:
```bash
openssl req -newkey rsa:4096 -x509 -sha256 -days 3650 -nodes -out tls/ota_server.crt -keyout tls/ota_server.key
openssl x509 -in tls/ota_server.crt -out tls/ota_server.pem -outform PEM
rm tls/ota_server.crt
```

# Swag
```bash
swag init -g http_api.go
```
