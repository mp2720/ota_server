package main

//	@title		OTA server
//	@version	1.0

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-Token

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"

	_ "mp1884/ota_server/docs"
)

type Api struct {
	firmwareSvc *FirmwareService
	tokenSvc    *TokenService
	cfg         *Config
}

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ApiFirmwareInfoResponse struct {
	Id          int64  `json:"id"`
	RepoName    string `json:"repo_name"`
	CommitId    string `json:"commit_id"`
	Tag         string `json:"tag"`
	BuiltAt     int64  `json:"built_at"`
	LoadedAt    int64  `json:"loaded_at"`
	LoadedBy    string `json:"loaded_by"`
	Sha256      string `json:"sha256"`
	Description string `json:"description"`
}

type ApiFirmwareResponse struct {
	Info   ApiFirmwareInfoResponse `json:"info"`
	BinUrl string                  `json:"bin_url"`
}

type ApiAddFirmwareInfoRequest struct {
	RepoName    string `json:"repo_name" binding:"required"`
	CommitId    string `json:"commit_id" binding:"required"`
	Tag         string `json:"tag" binding:"alphanum"`
	BuiltAt     int64  `json:"built_at" binding:"required"`
	Sha256      string `json:"sha256"`
	Description string `json:"description"`
}

type ApiAddFirmwareRequest struct {
	Info      ApiAddFirmwareInfoRequest `json:"info" binding:"required"`
	BinBase64 string                    `json:"bin_base64" binding:"required"`
}

func (api *Api) newFirmwareResponse(info *FirmwareInfo) ApiFirmwareResponse {
	return ApiFirmwareResponse{
		ApiFirmwareInfoResponse{
			info.Id,
			info.RepoName,
			info.CommitId,
			info.Tag,
			info.BuiltAt.Unix(),
			info.LoadedAt.Unix(),
			info.LoadedBy,
			info.Sha256,
			info.Description,
		},
		fmt.Sprintf("%s/bin/%d", api.cfg.host, info.Id),
	}
}

func (api *Api) auth(c *gin.Context, canBeBoard bool) (*TokenSubject, bool) {
	token := c.GetHeader("X-Token")
	subject, err := api.tokenSvc.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, HttpError{
			http.StatusUnauthorized,
			err.Error(),
		})
		return nil, false
	}

	if !canBeBoard && subject.isBoard {
		c.JSON(http.StatusForbidden, HttpError{
			http.StatusForbidden,
			"access denied",
		})
		return nil, false
	}

	return subject, true
}

// getNewestFirmware godoc
//
//	@Summary	Get latest firmware version
//	@Schemes
//	@Description	Get latest firmware version for given repo and tags. Available for boards.
//	@Produce		json
//	@Param			repo	query		string				false	"name of firmware's repo"
//	@Param			tags	query		array				false	"one of tags that firmware should have, can be omitted"
//	@Success		200		{object}	ApiFirmwareResponse	"ok"
//	@Failure		401		{object}	HttpError			"Invalid auth token"
//	@Failure		404		{object}	HttpError			"firmware for given repo and tags not found"
//	@Security		ApiKeyAuth
//	@Router			/firmwares/latest [get]
func (api *Api) getNewestFirmware(c *gin.Context) {
	_, ok := api.auth(c, true)
	if !ok {
		return
	}

	repoName := c.Query("repo")
	tagsStr := c.Query("tags")
	tags := strings.Split(tagsStr, ",")
	if len(tags) == 1 && tags[0] == "" {
		tags = []string{}
	}

	fmt.Printf("repo=%s tags=%+v\n", repoName, tags)

	fi, err := api.firmwareSvc.GetNewestFirmware(repoName, tags)
	if err != nil {
		panic(err)
	}

	if fi == nil {
		c.JSON(http.StatusNotFound, HttpError{
			http.StatusNotFound,
			"no firmware found for given repo and tags",
		})
		return
	}

	c.JSON(http.StatusOK, api.newFirmwareResponse(fi))
}

// getAllFirmwares godoc
//
//	@Summary	Get all firmwares
//	@Schemes
//	@Description	Get all firmwares
//	@Produce		json
//	@Success		200	{array}		ApiFirmwareResponse	"ok"
//	@Failure		401	{object}	HttpError			"Invalid auth token"
//	@Failure		403	{object}	HttpError			"Access is denied"
//	@Security		ApiKeyAuth
//	@Router			/firmwares [get]
func (api *Api) getAllFirmwares(c *gin.Context) {
	_, ok := api.auth(c, false)
	if !ok {
		return
	}

	fis, err := api.firmwareSvc.GetFirmwaresInfo()
	if err != nil {
		panic(err)
	}

	firmwares := []ApiFirmwareResponse{}
	for _, fi := range fis {
		firmwares = append(firmwares, api.newFirmwareResponse(&fi))
	}

	c.JSON(http.StatusOK, firmwares)
}

// addFirmware godoc
//
//	@Summary	Add firmware information and binary
//	@Schemes
//	@Accept			json
//	@Description	Get all firmwares
//	@Produce		json
//	@Param			firmware	body		ApiAddFirmwareRequest	true	"firmware info and binary"
//	@Success		201			{object}	ApiFirmwareResponse		"ok"
//	@Failure		401			{object}	HttpError				"Invalid auth token"
//	@Failure		403			{object}	HttpError				"Access is denied"
//	@Security		ApiKeyAuth
//	@Router			/firmwares [post]
func (api *Api) addFirmware(c *gin.Context) {
	subject, ok := api.auth(c, false)
	if !ok {
		return
	}

	var json ApiAddFirmwareRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, HttpError{
			http.StatusBadRequest,
			err.Error(),
		})
		return
	}

	info := FirmwareInfo{
		RepoName:    json.Info.RepoName,
		CommitId:    json.Info.CommitId,
		Tag:         json.Info.Tag,
		BuiltAt:     time.Unix(json.Info.BuiltAt, 0),
		LoadedBy:    subject.name,
		LoadedAt:    time.Now(),
		Sha256:      json.Info.Sha256,
		Description: json.Info.Description,
	}

	bytes, err := base64.StdEncoding.DecodeString(json.BinBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, HttpError{
			http.StatusBadRequest,
			err.Error(),
		})
		return
	}

	addedInfo, err := api.firmwareSvc.AddFirmware(&info, bytes)
	if err != nil {
		switch err.(type) {
		case *SHA256DiffersError:
			c.JSON(http.StatusBadRequest, HttpError{
				http.StatusBadRequest,
				err.Error(),
			})
			return
		default:
			panic(err)
		}
	}

	c.JSON(http.StatusCreated, api.newFirmwareResponse(addedInfo))
}

// getBinFile godoc
//
//	@Summary	Get binary file
//	@Schemes
//	@Description	Get binary firmware file with given id
//	@Param			id	path		int	true	"firmware's ID"
//	@Success		200	{file}		file
//	@Failure		401	{object}	HttpError	"Invalid auth token"
//	@Failure		403	{object}	HttpError	"Access is denied"
//	@Failure		404	{object}	HttpError	"firmware for given repo and tags not found"
//	@Security		ApiKeyAuth
//	@Router			/bin/{id} [get]
func (api *Api) getBinFile(c *gin.Context) {
	_, ok := api.auth(c, true)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, HttpError{
			http.StatusBadRequest,
			"id is invalid integer",
		})
		return
	}

	path, err := api.firmwareSvc.GetFirmwareBinaryPath(id)
	if err != nil {
		panic(err)
	}

	if path == "" {
		c.JSON(http.StatusNotFound, HttpError{
			http.StatusNotFound,
			"firmware not found",
		})
	}

	c.File(path)
}

func (api *Api) StartServer() error {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/firmwares/latest", api.getNewestFirmware)
		v1.GET("/firmwares", api.getAllFirmwares)
		v1.POST("/firmwares", api.addFirmware)
		v1.GET("/bin/:id", api.getBinFile)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run()
}
