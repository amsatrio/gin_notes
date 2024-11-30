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

type MBiodataService interface {
	GetMBiodata(context context.Context, id uint) (*model.MBiodata, error)
	CreateMBiodata(context context.Context, mBiodata *model.MBiodata, mUser *model.MUser) error
	UpdateMBiodata(context context.Context, mBiodata *model.MBiodata, mUser *model.MUser) error
	DeleteMBiodata(context context.Context, id uint, mUser *model.MUser) error
	SoftDeleteMBiodata(context context.Context, id uint, mUser *model.MUser) error
	GetPageMBiodata(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type MBiodataServiceImpl struct {
	db *gorm.DB
}

func NewMBiodataServiceImpl(db *gorm.DB) MBiodataService {
	return &MBiodataServiceImpl{
		db: db,
	}
}

func (s *MBiodataServiceImpl) GetMBiodata(context context.Context, id uint) (*model.MBiodata, error) {
	mBiodata := model.MBiodata{}
	result := s.db.First(&mBiodata, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mBiodata, nil
}

func (s *MBiodataServiceImpl) CreateMBiodata(context context.Context, mBiodata *model.MBiodata, mUser *model.MUser) error {

	// get id creator

	mBiodata.CreatedOn = response.JSONTime{Time: time.Now()}
	mBiodata.CreatedBy = mUser.Id

	util.Log("INFO", "service", "MBiodataService", "CreateMBiodata: ")

	var oldMBiodata model.MBiodata

	// find data
	result := s.db.First(&oldMBiodata, mBiodata.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&mBiodata)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MBiodataServiceImpl) UpdateMBiodata(context context.Context, mBiodata *model.MBiodata, mUser *model.MUser) error {

	var oldMBiodata *model.MBiodata

	// find data
	result := s.db.First(&oldMBiodata, mBiodata.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldMBiodata.Fullname = mBiodata.Fullname
	oldMBiodata.MobilePhone = mBiodata.MobilePhone
	oldMBiodata.Image = mBiodata.Image
	oldMBiodata.ImagePath = mBiodata.ImagePath
	oldMBiodata.ModifiedBy = mUser.Id
	oldMBiodata.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*mBiodata = *oldMBiodata

	result = s.db.Model(&oldMBiodata).Updates(oldMBiodata)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MBiodataServiceImpl) DeleteMBiodata(context context.Context, id uint, mUser *model.MUser) error {
	var mBiodata model.MBiodata
	result := s.db.Delete(&mBiodata, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *MBiodataServiceImpl) SoftDeleteMBiodata(context context.Context, id uint, mUser *model.MUser) error {
	var oldMBiodata = &model.MBiodata{}

	// find data
	result := s.db.First(&oldMBiodata, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	mBiodata := oldMBiodata
	mBiodata.DeletedOn = response.JSONTime{Time: time.Now()}
	mBiodata.DeletedBy = mUser.Id
	bool_true := true
	mBiodata.IsDelete = &bool_true

	result = s.db.Model(&oldMBiodata).Updates(mBiodata)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MBiodataServiceImpl) GetPageMBiodata(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageMBiodata", "")

	var mBiodatas []model.MBiodata
	var mBiodata model.MBiodata
	mBiodataMap := util.GetJSONFieldTypes(mBiodata)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, mBiodataMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&mBiodatas).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&mBiodatas)

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
		Content:          mBiodatas,
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

	util.Log("INFO", "service", "GetPageMBiodata", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
