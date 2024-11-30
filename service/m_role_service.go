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

type MRoleService interface {
	GetMRole(context context.Context, id uint) (*model.MRole, error)
	CreateMRole(context context.Context, mRole *model.MRole, mUser *model.MUser) error
	UpdateMRole(context context.Context, mRole *model.MRole, mUser *model.MUser) error
	DeleteMRole(context context.Context, id uint, mUser *model.MUser) error
	SoftDeleteMRole(context context.Context, id uint, mUser *model.MUser) error
	GetPageMRole(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type MRoleServiceImpl struct {
	db *gorm.DB
}

func NewMRoleServiceImpl(db *gorm.DB) MRoleService {
	return &MRoleServiceImpl{
		db: db,
	}
}

func (s *MRoleServiceImpl) GetMRole(context context.Context, id uint) (*model.MRole, error) {
	mRole := model.MRole{}
	result := s.db.First(&mRole, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mRole, nil
}

func (s *MRoleServiceImpl) CreateMRole(context context.Context, mRole *model.MRole, mUser *model.MUser) error {

	// get id creator

	mRole.CreatedOn = response.JSONTime{Time: time.Now()}
	mRole.CreatedBy = mUser.Id

	util.Log("INFO", "service", "MRoleService", "CreateMRole: ")

	var oldMRole model.MRole

	// find data
	result := s.db.First(&oldMRole, mRole.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&mRole)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MRoleServiceImpl) UpdateMRole(context context.Context, mRole *model.MRole, mUser *model.MUser) error {

	var oldMRole *model.MRole

	// find data
	result := s.db.First(&oldMRole, mRole.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldMRole.Name = mRole.Name
	oldMRole.Code = mRole.Code
	oldMRole.Level = mRole.Level
	oldMRole.ModifiedBy = mUser.Id
	oldMRole.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*mRole = *oldMRole

	result = s.db.Model(&oldMRole).Updates(oldMRole)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MRoleServiceImpl) DeleteMRole(context context.Context, id uint, mUser *model.MUser) error {
	var mRole model.MRole
	result := s.db.Delete(&mRole, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *MRoleServiceImpl) SoftDeleteMRole(context context.Context, id uint, mUser *model.MUser) error {
	var oldMRole = &model.MRole{}

	// find data
	result := s.db.First(&oldMRole, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	mRole := oldMRole
	mRole.DeletedOn = response.JSONTime{Time: time.Now()}
	mRole.DeletedBy = mUser.Id
	bool_true := true
	mRole.IsDelete = &bool_true

	result = s.db.Model(&oldMRole).Updates(mRole)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MRoleServiceImpl) GetPageMRole(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageMRole", "")

	var mRoles []model.MRole
	var mRole model.MRole
	mRoleMap := util.GetJSONFieldTypes(mRole)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, mRoleMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&mRoles).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&mRoles)

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
		Content:          mRoles,
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

	util.Log("INFO", "service", "GetPageMRole", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
