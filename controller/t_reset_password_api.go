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

// TResetPasswordPage godoc
//
//	@Summary		TResetPasswordPage
//	@Description	Get Page TResetPassword
//	@Tags			tResetPassword
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
//	@Router			/v1/t_reset_password [get]
func TResetPasswordPage(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "TResetPasswordPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}
	var filters []request.Filter
	jsonUnmarshalErr = json.Unmarshal([]byte(filterRequest), &filters)
	if jsonUnmarshalErr != nil {
		util.Log("ERROR", "controllers", "TResetPasswordPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)
	result, err := tResetPasswordService.GetPageTResetPassword(
		c,
		sorts,
		filters,
		searchRequest,
		pageInt,
		sizeInt64,
		sizeInt)

	if err != nil {
		util.Log("ERROR", "controllers", "TResetPasswordPage", "error: "+err.Error())
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

// TResetPasswordCreate godoc
//
//	@Summary		TResetPasswordCreate
//	@Description	Create TResetPassword
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			tResetPassword	body		model.TResetPassword	true	"Add TResetPassword"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password [post]
func TResetPasswordCreate(c *gin.Context) {

	// get request body
	body := model.TResetPassword{}

	// validate
	err := c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "TResetPasswordCreate", "bind error: "+err.Error(), err)
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
		util.Log("ERROR", "controllers", "TResetPasswordCreate", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)

	err = tResetPasswordService.CreateTResetPassword(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "TResetPasswordCreate", "create error: "+err.Error())
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

// TResetPasswordUpdate godoc
//
//	@Summary		TResetPasswordUpdate
//	@Description	Update TResetPassword
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			tResetPassword	body		model.TResetPassword	true	"Update TResetPassword"
//	@Param			id	path		int	true	"TResetPassword id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password/{id} [put]
func TResetPasswordUpdate(c *gin.Context) {

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

	body := model.TResetPassword{}

	// validate
	err = c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "TResetPasswordUpdate", "bind error: "+err.Error(), err)
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
		util.Log("ERROR", "controllers", "TResetPasswordUpdate", "error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)

	err = tResetPasswordService.UpdateTResetPassword(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "TResetPasswordUpdate UpdateTResetPassword", err.Error())
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

// TResetPasswordIndex godoc
//
//	@Summary		TResetPasswordIndex
//	@Description	Get TResetPassword by id
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"TResetPassword id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password/{id} [get]
func TResetPasswordIndex(c *gin.Context) {

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

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)

	tResetPassword, err := tResetPasswordService.GetTResetPassword(c, idUint)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Set(constant.ERROR_KEY, constant.ErrorDataNotFound)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	if err != nil {
		util.Log("ERROR", "controllers", "TResetPasswordIndex", err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRetrieveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = tResetPassword
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// TResetPasswordDelete godoc
//
//	@Summary		TResetPasswordDelete
//	@Description	Delete TResetPassword by id
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"TResetPassword id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password/{id} [delete]
func TResetPasswordDelete(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "TResetPasswordDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)

	// delete tResetPassword
	err = tResetPasswordService.DeleteTResetPassword(c, idUint, mUser)

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

// TResetPasswordSoftDelete godoc
//
//	@Summary		TResetPasswordSoftDelete
//	@Description	Soft Delete TResetPassword by id
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"TResetPassword id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password/delete/{id} [put]
func TResetPasswordSoftDelete(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "TResetPasswordSoftDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	tResetPasswordService := service.NewTResetPasswordServiceImpl(initializer.DB)

	// delete tResetPassword
	err = tResetPasswordService.SoftDeleteTResetPassword(c, idUint, mUser)

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

// TResetPasswordHeader godoc
//
//	@Summary		TResetPasswordHeader
//	@Description	Get TResetPassword header
//	@Tags			tResetPassword
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/t_reset_password/header [get]
func TResetPasswordHeader(c *gin.Context) {
	header := util.GetJSONFieldTypes(model.TResetPassword{})

	// return response
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = header
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}
