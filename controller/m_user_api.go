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

// MUserPage godoc
//
//	@Summary		MUserPage
//	@Description	Get Page MUser
//	@Tags			mUser
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
//	@Router			/v1/m_user [get]
func MUserPage(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "MUserPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}
	var filters []request.Filter
	jsonUnmarshalErr = json.Unmarshal([]byte(filterRequest), &filters)
	if jsonUnmarshalErr != nil {
		util.Log("ERROR", "controllers", "MUserPage", "jsonUnmarshalErr error: "+jsonUnmarshalErr.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, jsonUnmarshalErr)
		c.Abort()
		return
	}

	mUserService := service.NewMUserServiceImpl(initializer.DB)
	result, err := mUserService.GetPageMUser(
		c,
		sorts,
		filters,
		searchRequest,
		pageInt,
		sizeInt64,
		sizeInt)

	if err != nil {
		util.Log("ERROR", "controllers", "MUserPage", "error: "+err.Error())
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

// MUserCreate godoc
//
//	@Summary		MUserCreate
//	@Description	Create MUser
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			mUser	body		model.MUser	true	"Add MUser"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user [post]
func MUserCreate(c *gin.Context) {

	// get request body
	body := model.MUser{}

	// validate
	err := c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "MUserCreate", "bind error: "+err.Error(), err)
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
		util.Log("ERROR", "controllers", "MUserCreate", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	err = mUserService.CreateMUser(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "MUserCreate", "create error: "+err.Error())
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

// MUserUpdate godoc
//
//	@Summary		MUserUpdate
//	@Description	Update MUser
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			mUser	body		model.MUser	true	"Update MUser"
//	@Param			id	path		int	true	"MUser id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user/{id} [put]
func MUserUpdate(c *gin.Context) {

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

	body := model.MUser{}

	// validate
	err = c.ShouldBindJSON(&body)
	if err != nil {
		util.LogError("controllers", "MUserUpdate", "bind error: "+err.Error(), err)
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
		util.Log("ERROR", "controllers", "MUserUpdate", "error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	err = mUserService.UpdateMUser(c, &body, mUser)

	if err != nil {
		util.Log("ERROR", "controllers", "MUserUpdate UpdateMUser", err.Error())
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

// MUserIndex godoc
//
//	@Summary		MUserIndex
//	@Description	Get MUser by id
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MUser id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user/{id} [get]
func MUserIndex(c *gin.Context) {

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

	mUserService := service.NewMUserServiceImpl(initializer.DB)

	mUser, err := mUserService.GetMUser(c, idUint)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.Set(constant.ERROR_KEY, constant.ErrorDataNotFound)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	if err != nil {
		util.Log("ERROR", "controllers", "MUserIndex", err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRetrieveDataFailed)
		c.Set(constant.ERROR_MESSAGE, err)
		c.Abort()
		return
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = mUser
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

// MUserDelete godoc
//
//	@Summary		MUserDelete
//	@Description	Delete MUser by id
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MUser id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user/{id} [delete]
func MUserDelete(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "MUserDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// delete mUser
	err = mUserService.DeleteMUser(c, idUint, mUser)

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

// MUserSoftDelete godoc
//
//	@Summary		MUserSoftDelete
//	@Description	Soft Delete MUser by id
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Param			id	path		int	true	"MUser id"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user/delete/{id} [put]
func MUserSoftDelete(c *gin.Context) {
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
		util.Log("ERROR", "controllers", "MUserSoftDelete", "create error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorUserNotFound)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// delete mUser
	err = mUserService.SoftDeleteMUser(c, idUint, mUser)

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

// MUserHeader godoc
//
//	@Summary		MUserHeader
//	@Description	Get MUser header
//	@Tags			mUser
//	@Accept			json
//	@Produce		json
//	@Param			Accept-Encoding	header	string	false	"gzip" default(gzip)
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/v1/m_user/header [get]
func MUserHeader(c *gin.Context) {
	header := util.GetJSONFieldTypes(model.MUser{})

	// return response
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = header
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}
