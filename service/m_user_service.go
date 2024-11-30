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

type MUserService interface {
	GetMUser(context context.Context, id uint) (*model.MUser, error)
	GetMUserByEmail(context context.Context, email string) (*model.MUser, error)
	CreateMUser(context context.Context, mUser *model.MUser, mUserAccess *model.MUser) error
	UpdateMUser(context context.Context, mUser *model.MUser, mUserAccess *model.MUser) error
	DeleteMUser(context context.Context, id uint, mUserAccess *model.MUser) error
	SoftDeleteMUser(context context.Context, id uint, mUserAccess *model.MUser) error
	GetPageMUser(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type MUserServiceImpl struct {
	db *gorm.DB
}

func NewMUserServiceImpl(db *gorm.DB) MUserService {
	return &MUserServiceImpl{
		db: db,
	}
}

func (s *MUserServiceImpl) GetMUser(context context.Context, id uint) (*model.MUser, error) {
	mUser := model.MUser{}
	result := s.db.First(&mUser, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mUser, nil
}

func (s *MUserServiceImpl) GetMUserByEmail(context context.Context, email string) (*model.MUser, error) {
	mUser := model.MUser{}
	result := s.db.Where("email = ?", email).First(&mUser)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mUser, nil
}

func (s *MUserServiceImpl) CreateMUser(context context.Context, mUser *model.MUser, mUserAccess *model.MUser) error {

	// get id creator

	mUser.CreatedOn = response.JSONTime{Time: time.Now()}
	mUser.CreatedBy = mUserAccess.Id

	util.Log("INFO", "service", "MUserService", "CreateMUser: ")

	var oldMUser model.MUser

	// find data
	result := s.db.First(&oldMUser, mUser.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&mUser)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MUserServiceImpl) UpdateMUser(context context.Context, mUser *model.MUser, mUserAccess *model.MUser) error {

	var oldMUser *model.MUser

	// find data
	result := s.db.First(&oldMUser, mUser.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldMUser.BiodataId = mUser.BiodataId
	oldMUser.RoleId = mUser.RoleId
	oldMUser.Email = mUser.Email
	oldMUser.LoginAttempt = mUser.LoginAttempt
	oldMUser.Password = mUser.Password
	oldMUser.IsLocked = mUser.IsLocked
	oldMUser.LastLogin = mUser.LastLogin
	oldMUser.ModifiedBy = mUserAccess.Id
	oldMUser.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*mUser = *oldMUser

	result = s.db.Model(&oldMUser).Updates(oldMUser)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MUserServiceImpl) DeleteMUser(context context.Context, id uint, mUserAccess *model.MUser) error {
	var mUser model.MUser
	result := s.db.Delete(&mUser, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *MUserServiceImpl) SoftDeleteMUser(context context.Context, id uint, mUserAccess *model.MUser) error {
	var oldMUser = &model.MUser{}

	// find data
	result := s.db.First(&oldMUser, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	mUser := oldMUser
	mUser.DeletedOn = response.JSONTime{Time: time.Now()}
	mUser.DeletedBy = mUser.Id
	bool_true := true
	mUser.IsDelete = &bool_true

	result = s.db.Model(&oldMUser).Updates(mUser)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MUserServiceImpl) GetPageMUser(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageMUser", "")

	var mUsers []model.MUser
	var mUser model.MUser
	mUserMap := util.GetJSONFieldTypes(mUser)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, mUserMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&mUsers).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&mUsers)

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
		Content:          mUsers,
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

	util.Log("INFO", "service", "GetPageMUser", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
