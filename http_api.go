package main

//	@title		OTA server
//	@version	1.0

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						X-Token

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

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
	Id          int64    `json:"id"`
	RepoName    string   `json:"repo_name"`
	CommitId    string   `json:"commit_id"`
	Boards      []string `json:"boards"`
	CreatedAt   int64    `json:"created_at"`
	LoadedBy    string   `json:"loaded_by"`
	Md5         string   `json:"md5"`
	Description string   `json:"description"`
	Size        int      `json:"size"`
}

type ApiFirmwareResponse struct {
	Info   ApiFirmwareInfoResponse `json:"info"`
	BinUrl string                  `json:"bin_url"`
}

type ApiAddFirmwareInfoRequest struct {
	RepoName    string   `json:"repo_name" binding:"required"`
	CommitId    string   `json:"commit_id"`
	Boards      []string `json:"boards" binding:"required,min=1,dive,min=1"`
	Description string   `json:"description"`
}

type ApiUserResponse struct {
	Name    string `json:"name"`
	IsBoard bool   `json:"is_board"`
}

func (api *Api) newFirmwareResponse(info *FirmwareInfo) ApiFirmwareResponse {
	var binUrl string
	if info.hasBin() {
		binUrl = fmt.Sprintf("%s/bin/%d", api.cfg.host, info.Id)
	} else {
		binUrl = ""
	}

	return ApiFirmwareResponse{
		ApiFirmwareInfoResponse{
			info.Id,
			info.RepoName,
			info.CommitId,
			info.Boards,
			info.CreatedAt.Unix(),
			info.LoadedBy,
			info.Md5,
			info.Description,
			info.Size,
		},
		binUrl,
	}
}

func (api *Api) auth(c *gin.Context, constraints *TokenSubject) (*TokenSubject, bool) {
	token := c.GetHeader("X-Token")
	subject, err := api.tokenSvc.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, HttpError{
			http.StatusUnauthorized,
			err.Error(),
		})
		return nil, false
	}

	if constraints != nil && constraints.isBoard != subject.isBoard {
		c.JSON(http.StatusForbidden, HttpError{
			http.StatusForbidden,
			"access denied",
		})
		return nil, false
	}

	return subject, true
}

// getLatestFirmware godoc
//
//	@Summary	Get latest firmware version
//	@Schemes
//	@Description	Get latest firmware version for given repo and tags. Only for boards
//	@Produce		json
//	@Param			repo	query		string				false	"name of firmware's repo"
//	@Success		200		{object}	ApiFirmwareResponse	"ok"
//	@Failure		401		{object}	HttpError			"Invalid auth token"
//	@Failure		403		{object}	HttpError			"Access is denied"
//	@Failure		404		{object}	HttpError			"no firmware found for this board in repo"
//	@Security		ApiKeyAuth
//	@Router			/firmwares/latest [get]
func (api *Api) getLatestFirmware(c *gin.Context) {
	subject, ok := api.auth(c, &TokenSubject{isBoard: true})
	if !ok {
		return
	}

	fi, err := api.firmwareSvc.GetLatestFirmware(c.Query("repo"), subject.name)
	if err != nil {
		panic(err)
	}

	if fi == nil {
		c.JSON(http.StatusNotFound, HttpError{
			http.StatusNotFound,
			"no firmware found for this board in repo",
		})
		return
	}

	c.JSON(http.StatusOK, api.newFirmwareResponse(fi))
}

// getAllFirmwares godoc
//
//	@Summary	Get all firmwares
//	@Schemes
//	@Description	Get all firmwares. Only for non-board users
//	@Produce		json
//	@Success		200	{array}		ApiFirmwareResponse	"ok"
//	@Failure		401	{object}	HttpError			"Invalid auth token"
//	@Failure		403	{object}	HttpError			"Access is denied"
//	@Security		ApiKeyAuth
//	@Router			/firmwares [get]
func (api *Api) getAllFirmwares(c *gin.Context) {
	_, ok := api.auth(c, &TokenSubject{isBoard: false})
	if !ok {
		return
	}

	fis, err := api.firmwareSvc.GetAllFirmwaresInfo()
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
//	@Summary	Create firmware record in db
//	@Schemes
//	@Accept			json
//	@Description	Create firmare record in db. Upload file to POST /bin/{id} after. Only for non-board users
//	@Produce		json
//	@Param			firmware	body		ApiAddFirmwareInfoRequest	true	"firmware info"
//	@Success		201			{object}	ApiFirmwareResponse			"ok"
//	@Failure		401			{object}	HttpError					"Invalid auth token"
//	@Failure		403			{object}	HttpError					"Access is denied"
//	@Security		ApiKeyAuth
//	@Router			/firmwares [post]
func (api *Api) addFirmware(c *gin.Context) {
	subject, ok := api.auth(c, &TokenSubject{isBoard: false})
	if !ok {
		return
	}

	var json ApiAddFirmwareInfoRequest
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, HttpError{
			http.StatusBadRequest,
			err.Error(),
		})
		return
	}

	info := FirmwareInfo{
		RepoName:    json.RepoName,
		CommitId:    json.CommitId,
		Boards:      json.Boards,
		CreatedAt:   time.Now(),
		LoadedBy:    subject.name,
		Description: json.Description,
	}

	addedInfo, err := api.firmwareSvc.CreateFirmware(&info)
	if err != nil {
		switch err.(type) {
		case *Md5DiffersError:
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

// getFirmwareBinary godoc
//
//	@Summary	Get binary file
//	@Schemes
//	@Description	Get binary firmware file with given id. Available for all authenticated users
//	@Param			id	path		int	true	"firmware's ID"
//	@Success		200	{file}		file
//	@Failure		401	{object}	HttpError	"Invalid auth token"
//	@Failure		404	{object}	HttpError	"firmware not found"
//	@Security		ApiKeyAuth
//	@Router			/bin/{id} [get]
func (api *Api) getFirmwareBinary(c *gin.Context) {
	_, ok := api.auth(c, nil)
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

// getAuthenticatedUser godoc
//
//	@Summary	Get authenticated user
//	@Schemes
//	@Produce		json
//	@Description	Get authenticated user
//	@Success		200	{object}	ApiUserResponse	"ok"
//	@Failure		401	{object}	HttpError		"Invalid auth token"
//	@Security		ApiKeyAuth
//	@Router			/users/me [get]
func (api *Api) getAuthenticatedUser(c *gin.Context) {
	subj, ok := api.auth(c, nil)
	if !ok {
		return
	}

	c.JSON(http.StatusOK, ApiUserResponse{
		subj.name,
		subj.isBoard,
	})
}

// addFirmwareBinary godoc
//
//	@Schemes
//	@Produce		json
//	@Summary		Upload firmware binary file
//	@Description	Upload firmware binary file. Only for non-board users
//	@Accept			multipart/form-data
//	@Param			id		path		int		true	"firmware's ID"
//	@Param			file	formData	file	true	"firmware binary file"
//	@Success		204
//	@Failure		400	{object}	HttpError	"File is already uploaded/empty file provided/invalid id"
//	@Failure		401	{object}	HttpError	"Invalid auth token"
//	@Failure		403	{object}	HttpError	"Access denied"
//	@Failure		404	{object}	HttpError	"Firmware not found"
//	@Security		ApiKeyAuth
//	@Router			/bin/{id} [post]
func (api *Api) addFirmwareBinary(c *gin.Context) {
	_, ok := api.auth(c, &TokenSubject{isBoard: false})
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

	fh, err := c.FormFile("file")
	if err != nil {
        c.JSON(http.StatusBadRequest, HttpError{
            http.StatusBadRequest,
            err.Error(),
        })
	}

	if fh.Size == 0 {
		c.JSON(http.StatusBadRequest, HttpError{
			http.StatusBadRequest,
			"empty file is not allowed",
		})
		return
	}

	f, err := fh.Open()
	if err != nil {
		panic(err)
	}

	bytes := make([]byte, fh.Size)
	f.Read(bytes)
	if err := api.firmwareSvc.AddFirmwareFile(id, bytes); err != nil {
		switch err.(type) {
		case *FirmwareNotFoundError:
			c.JSON(http.StatusNotFound, HttpError{
				http.StatusNotFound,
				"firmware not found",
			})
			return
		case *FirmwareFileAlreadyUploaded:
			c.JSON(http.StatusBadRequest, HttpError{
				http.StatusBadRequest,
				"file is already uploaded",
			})
			return
		default:
			panic(err)
		}
	}

	c.Status(http.StatusNoContent)
}

func (api *Api) StartServer() error {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.GET("/firmwares/latest", api.getLatestFirmware)
		v1.GET("/firmwares", api.getAllFirmwares)
		v1.POST("/firmwares", api.addFirmware)
		v1.GET("/bin/:id", api.getFirmwareBinary)
		v1.POST("/bin/:id", api.addFirmwareBinary)
		v1.GET("/users/me", api.getAuthenticatedUser)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.Run(api.cfg.port)
}
