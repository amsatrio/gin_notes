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

type TTokenService interface {
	GetTToken(context context.Context, id uint) (*model.TToken, error)
	CreateTToken(context context.Context, tToken *model.TToken, mUser *model.MUser) error
	UpdateTToken(context context.Context, tToken *model.TToken, mUser *model.MUser) error
	DeleteTToken(context context.Context, id uint, mUser *model.MUser) error
	SoftDeleteTToken(context context.Context, id uint, mUser *model.MUser) error
	GetPageTToken(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type TTokenServiceImpl struct {
	db *gorm.DB
}

func NewTTokenServiceImpl(db *gorm.DB) TTokenService {
	return &TTokenServiceImpl{
		db: db,
	}
}

func (s *TTokenServiceImpl) GetTToken(context context.Context, id uint) (*model.TToken, error) {
	tToken := model.TToken{}
	result := s.db.First(&tToken, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &tToken, nil
}

func (s *TTokenServiceImpl) CreateTToken(context context.Context, tToken *model.TToken, mUser *model.MUser) error {

	// get id creator

	tToken.CreatedOn = response.JSONTime{Time: time.Now()}
	tToken.CreatedBy = mUser.Id

	util.Log("INFO", "service", "TTokenService", "CreateTToken: ")

	var oldTToken model.TToken

	// find data
	result := s.db.First(&oldTToken, tToken.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&tToken)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TTokenServiceImpl) UpdateTToken(context context.Context, tToken *model.TToken, mUser *model.MUser) error {

	var oldTToken *model.TToken

	// find data
	result := s.db.First(&oldTToken, tToken.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldTToken.Email = tToken.Email
	oldTToken.UserId = tToken.UserId
	oldTToken.Token = tToken.Token
	oldTToken.ExpiredOn = tToken.ExpiredOn
	oldTToken.IsExpired = tToken.IsExpired
	oldTToken.UsedFor = tToken.UsedFor
	oldTToken.ModifiedBy = mUser.Id
	oldTToken.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*tToken = *oldTToken

	result = s.db.Model(&oldTToken).Updates(oldTToken)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TTokenServiceImpl) DeleteTToken(context context.Context, id uint, mUser *model.MUser) error {
	var tToken model.TToken
	result := s.db.Delete(&tToken, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *TTokenServiceImpl) SoftDeleteTToken(context context.Context, id uint, mUser *model.MUser) error {
	var oldTToken = &model.TToken{}

	// find data
	result := s.db.First(&oldTToken, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	tToken := oldTToken
	tToken.DeletedOn = response.JSONTime{Time: time.Now()}
	tToken.DeletedBy = mUser.Id
	bool_true := true
	tToken.IsDelete = &bool_true

	result = s.db.Model(&oldTToken).Updates(tToken)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *TTokenServiceImpl) GetPageTToken(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageTToken", "")

	var tTokens []model.TToken
	var tToken model.TToken
	tTokenMap := util.GetJSONFieldTypes(tToken)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, tTokenMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&tTokens).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&tTokens)

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
		Content:          tTokens,
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

	util.Log("INFO", "service", "GetPageTToken", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
