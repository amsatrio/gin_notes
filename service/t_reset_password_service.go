package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/request"
	"github.com/amsatrio/gin_notes/model/response"
	"github.com/amsatrio/gin_notes/util"
)

type TResetPasswordService interface {
	GetTResetPassword(context context.Context, id uint) (*model.TResetPassword, error)
	CreateTResetPassword(context context.Context, tResetPassword *model.TResetPassword, mUser *model.MUser) error
	UpdateTResetPassword(context context.Context, tResetPassword *model.TResetPassword, mUser *model.MUser) error
	DeleteTResetPassword(context context.Context, id uint, mUser *model.MUser) error
	SoftDeleteTResetPassword(context context.Context, id uint, mUser *model.MUser) error
	GetPageTResetPassword(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type TResetPasswordServiceImpl struct {
	db *gorm.DB
}

func NewTResetPasswordServiceImpl(db *gorm.DB) TResetPasswordService {
	return &TResetPasswordServiceImpl{
		db: db,
	}
}

func (s *TResetPasswordServiceImpl) GetTResetPassword(context context.Context, id uint) (*model.TResetPassword, error) {
	tResetPassword := model.TResetPassword{}
	result := s.db.First(&tResetPassword, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &tResetPassword, nil
}

func (s *TResetPasswordServiceImpl) CreateTResetPassword(context context.Context, tResetPassword *model.TResetPassword, mUser *model.MUser) error {

	// get id creator

	tResetPassword.CreatedOn = response.JSONTime{Time: time.Now()}
	tResetPassword.CreatedBy = mUser.Id

	util.Log("INFO", "service", "TResetPasswordService", "CreateTResetPassword: ")

	var oldTResetPassword model.TResetPassword

	// find data
	result := s.db.First(&oldTResetPassword, tResetPassword.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&tResetPassword)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TResetPasswordServiceImpl) UpdateTResetPassword(context context.Context, tResetPassword *model.TResetPassword, mUser *model.MUser) error {

	var oldTResetPassword *model.TResetPassword

	// find data
	result := s.db.First(&oldTResetPassword, tResetPassword.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldTResetPassword.OldPassword = tResetPassword.OldPassword
	oldTResetPassword.NewPassword = tResetPassword.NewPassword
	oldTResetPassword.ResetFor = tResetPassword.ResetFor
	oldTResetPassword.ModifiedBy = mUser.Id
	oldTResetPassword.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*tResetPassword = *oldTResetPassword

	result = s.db.Model(&oldTResetPassword).Updates(oldTResetPassword)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TResetPasswordServiceImpl) DeleteTResetPassword(context context.Context, id uint, mUser *model.MUser) error {
	var tResetPassword model.TResetPassword
	result := s.db.Delete(&tResetPassword, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *TResetPasswordServiceImpl) SoftDeleteTResetPassword(context context.Context, id uint, mUser *model.MUser) error {
	var oldTResetPassword = &model.TResetPassword{}

	// find data
	result := s.db.First(&oldTResetPassword, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	tResetPassword := oldTResetPassword
	tResetPassword.DeletedOn = response.JSONTime{Time: time.Now()}
	tResetPassword.DeletedBy = mUser.Id
	bool_true := true
	tResetPassword.IsDelete = &bool_true

	result = s.db.Model(&oldTResetPassword).Updates(tResetPassword)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TResetPasswordServiceImpl) GetPageTResetPassword(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageTResetPassword", "")

	var tResetPasswords []model.TResetPassword
	var tResetPassword model.TResetPassword
	tResetPasswordMap := util.GetJSONFieldTypes(tResetPassword)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, tResetPasswordMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&tResetPasswords).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&tResetPasswords)

	if result.Error != nil {
		return nil, result.Error
	}

	lastPage := int64(pageInt) == totalPages-1
	firstPage := pageInt == 0

	// prepare page
	sort := response.Sort{
		Empty:    totalElements <= 0,
		Sorted:   true,
		Unsorted: false,
	}

	pageable := response.Pageable{
		Offset:     pageInt * sizeInt,
		PageNumber: pageInt,
		PageSize:   sizeInt,
		Paged:      true,
		UnPaged:    false,
		Sort:       sort,
	}

	page := response.Page{
		Content:          tResetPasswords,
		Pageable:         pageable,
		Sort:             sort,
		TotalPages:       totalPages,
		TotalElements:    totalElements,
		Size:             sizeInt,
		Number:           pageInt,
		NumberOfElements: sizeInt,
		Last:             lastPage,
		First:            firstPage,
		Empty:            sort.Empty,
	}

	util.Log("INFO", "service", "GetPageTResetPassword", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
