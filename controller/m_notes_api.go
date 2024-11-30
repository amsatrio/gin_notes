package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/request"
	"github.com/amsatrio/gin_notes/model/response"
	"github.com/amsatrio/gin_notes/service"
	"github.com/amsatrio/gin_notes/util"
)

// MNotesPage godoc
//
//	@Summary		MNotesPage
//	@Description	Get Page MNotes
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			_page	query		string	false	"page" default(0)
//	@Param			_size	query		string	false	"size" default(5)
//	@Param			_sort	query		string	false	"sort"
//	@Param			_filter	query		string	false	"filter"
//	@Param			_q	query		string	false	"global filter"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes [get]
func MNotesPage(c *gin.Context) {
	sortRequest := c.DefaultQuery("_sort", "[]")
	pageRequest := c.DefaultQuery("_page", "0")
	sizeRequest := c.DefaultQuery("_size", "10")
	filterRequest := c.DefaultQuery("_filter", "[]")
	searchRequest := c.DefaultQuery("_q", "")

	pageInt, errorPageInt := strconv.Atoi(pageRequest)
	sizeInt64, errorLimitInt64 := strconv.ParseInt(sizeRequest, 10, 64)
	sizeInt, errorLimitInt := strconv.Atoi(sizeRequest)

	if errorPageInt != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, errorPageInt.Error())
		c.Abort()
		return
	}
	if errorLimitInt64 != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, errorLimitInt64.Error())
		c.Abort()
		return
	}
	if errorLimitInt != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, errorLimitInt.Error())
		c.Abort()
		return
	}

	isLetterNumber := regexp.MustCompile(`^[a-zA-Z0-9\s]+$`).MatchString
	if !isLetterNumber(searchRequest) && searchRequest != "" {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, errors.New("global search must not contains special character"))
		c.Abort()
		return
	}

	var sorts []request.Sort
	jsonUnmarshalErr := json.Unmarshal([]byte(sortRequest), &sorts)
	if jsonUnmarshalErr != nil {
		util.Log("ERROR", "controllers", "MNotesPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}
	var filters []request.Filter
	jsonUnmarshalErr = json.Unmarshal([]byte(filterRequest), &filters)
	if jsonUnmarshalErr != nil {
		util.Log("ERROR", "controllers", "MNotesPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)
	result, err := mNotesService.GetPageMNotes(
		c,
		sorts,
		filters,
		searchRequest,
		pageInt,
		sizeInt64,
		sizeInt)

	if err != nil {
		util.Log("ERROR", "controllers", "MNotesPage", "error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRetrieveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = *result
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesCreate godoc
//
//	@Summary		MNotesCreate
//	@Description	Create MNotes
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			mNotes	body		model.MNotes	true	"Add MNotes"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes [post]
func MNotesCreate(c *gin.Context) {

	// get request body
	body := model.MNotes{}

	// validate
	err := c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "MNotesCreate", "bind error: "+err.Error(), err)
		out, _ := util.ValidateError(err)
		if out != nil {
			c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
			c.Set(constant.ERROR_MESSAGE, out)
			c.Abort()
			return
		}
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	email := c.GetString("username")

	// find mUser
	mUserService := service.NewMUserServiceImpl(initializer.DB)
	mUser, err := mUserService.GetMUserByEmail(c, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		util.Log("ERROR", "controllers", "MNotesCreate", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)

	err = mNotesService.CreateMNotes(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "MNotesCreate", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorSaveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = body
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesUpdate godoc
//
//	@Summary		MNotesUpdate
//	@Description	Update MNotes
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			mNotes	body		model.MNotes	true	"Update MNotes"
//	@Param			id	path		int	true	"MNotes id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes/{id} [put]
func MNotesUpdate(c *gin.Context) {

	id := c.Param("id")
	var idUint uint
	idUint64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}
	idUint = uint(idUint64)

	body := model.MNotes{}

	// validate
	err = c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "MNotesUpdate", "bind error: "+err.Error(), err)
		out, _ := util.ValidateError(err)
		if out != nil {
			c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
			c.Set(constant.ERROR_MESSAGE, out)
			c.Abort()
			return
		}
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	body.Id = idUint

	email := c.GetString("username")

	// find mUser
	mUserService := service.NewMUserServiceImpl(initializer.DB)
	mUser, err := mUserService.GetMUserByEmail(c, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		util.Log("ERROR", "controllers", "MNotesUpdate", "error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)

	err = mNotesService.UpdateMNotes(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "MNotesUpdate UpdateMNotes", err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorSaveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = body
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesIndex godoc
//
//	@Summary		MNotesIndex
//	@Description	Get MNotes by id
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MNotes id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes/{id} [get]
func MNotesIndex(c *gin.Context) {

	id := c.Param("id")
	var idUint uint
	idUint64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}
	idUint = uint(idUint64)

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)

	mNotes, err := mNotesService.GetMNotes(c, idUint)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Set(constant.ERROR_KEY, constant.ErrorDataNotFound)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	if err != nil {
		util.Log("ERROR", "controllers", "MNotesIndex", err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRetrieveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = mNotes
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesDelete godoc
//
//	@Summary		MNotesDelete
//	@Description	Delete MNotes by id
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MNotes id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes/{id} [delete]
func MNotesDelete(c *gin.Context) {
	// get id from request param
	idParam := c.Param("id")
	var idUint uint
	idUint64, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}
	idUint = uint(idUint64)

	email := c.GetString("username")

	// find mUser
	mUserService := service.NewMUserServiceImpl(initializer.DB)
	mUser, err := mUserService.GetMUserByEmail(c, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		util.Log("ERROR", "controllers", "MNotesDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)

	// delete mNotes
	err = mNotesService.DeleteMNotes(c, idUint, mUser)

	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorDataNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// return response
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = nil
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesSoftDelete godoc
//
//	@Summary		MNotesSoftDelete
//	@Description	Soft Delete MNotes by id
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MNotes id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes/delete/{id} [put]
func MNotesSoftDelete(c *gin.Context) {
	// get id from request param
	id := c.Param("id")
	var idUint uint
	idUint64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}
	idUint = uint(idUint64)

	email := c.GetString("username")

	mUserService := service.NewMUserServiceImpl(initializer.DB)
	mUser, err := mUserService.GetMUserByEmail(c, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		util.Log("ERROR", "controllers", "MNotesSoftDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	mNotesService := service.NewMNotesServiceImpl(initializer.DB)

	// delete mNotes
	err = mNotesService.SoftDeleteMNotes(c, idUint, mUser)

	// validate error
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	// return response
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = nil
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MNotesHeader godoc
//
//	@Summary		MNotesHeader
//	@Description	Get MNotes header
//	@Tags			mNotes
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_notes/header [get]
func MNotesHeader(c *gin.Context) {
	header := util.GetJSONFieldTypes(model.MNotes{})

	// return response
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = header
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}
