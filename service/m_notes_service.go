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

type MNotesService interface {
	GetMNotes(context context.Context, id uint) (*model.MNotes, error)
	CreateMNotes(context context.Context, mNotes *model.MNotes, mUser *model.MUser) error
	UpdateMNotes(context context.Context, mNotes *model.MNotes, mUser *model.MUser) error
	DeleteMNotes(context context.Context, id uint, mUser *model.MUser) error
	SoftDeleteMNotes(context context.Context, id uint, mUser *model.MUser) error
	GetPageMNotes(
		context context.Context,
		sortRequest []request.Sort,
		filterRequest []request.Filter,
		searchRequest string,
		pageInt int,
		sizeInt64 int64,
		sizeInt int) (*response.Page, error)
}

type MNotesServiceImpl struct {
	db *gorm.DB
}

func NewMNotesServiceImpl(db *gorm.DB) MNotesService {
	return &MNotesServiceImpl{
		db: db,
	}
}

func (s *MNotesServiceImpl) GetMNotes(context context.Context, id uint) (*model.MNotes, error) {
	mNotes := model.MNotes{}
	result := s.db.First(&mNotes, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &mNotes, nil
}

func (s *MNotesServiceImpl) CreateMNotes(context context.Context, mNotes *model.MNotes, mUser *model.MUser) error {

	// get id creator

	mNotes.CreatedOn = response.JSONTime{Time: time.Now()}
	mNotes.CreatedBy = mUser.Id

	util.Log("INFO", "service", "MNotesService", "CreateMNotes: ")

	var oldMNotes model.MNotes

	// find data
	result := s.db.First(&oldMNotes, mNotes.Id)
	if result.Error == nil {
		return errors.New("data exist")
	}

	result = s.db.Create(&mNotes)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MNotesServiceImpl) UpdateMNotes(context context.Context, mNotes *model.MNotes, mUser *model.MUser) error {

	var oldMNotes *model.MNotes

	// find data
	result := s.db.First(&oldMNotes, mNotes.Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	oldMNotes.Content = mNotes.Content
	oldMNotes.Title = mNotes.Title
	oldMNotes.ModifiedBy = mUser.Id
	oldMNotes.ModifiedOn = response.JSONTime{Time: time.Now()}

	// update data for response
	*mNotes = *oldMNotes

	result = s.db.Model(&oldMNotes).Updates(oldMNotes)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MNotesServiceImpl) DeleteMNotes(context context.Context, id uint, mUser *model.MUser) error {
	var mNotes model.MNotes
	result := s.db.Delete(&mNotes, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("data not found")
	}
	return nil
}

func (s *MNotesServiceImpl) SoftDeleteMNotes(context context.Context, id uint, mUser *model.MUser) error {
	var oldMNotes = &model.MNotes{}

	// find data
	result := s.db.First(&oldMNotes, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("data not found")
	}
	if result.Error != nil {
		return result.Error
	}

	// update data
	mNotes := oldMNotes
	mNotes.DeletedOn = response.JSONTime{Time: time.Now()}
	mNotes.DeletedBy = mUser.Id
	bool_true := true
	mNotes.IsDelete = &bool_true

	result = s.db.Model(&oldMNotes).Updates(mNotes)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *MNotesServiceImpl) GetPageMNotes(
	context context.Context,
	sortRequest []request.Sort,
	filterRequest []request.Filter,
	searchRequest string,
	pageInt int,
	sizeInt64 int64,
	sizeInt int) (*response.Page, error) {

	util.Log("INFO", "service", "GetPageMNotes", "")

	var mNotess []model.MNotes
	var mNotes model.MNotes
	mNotesMap := util.GetJSONFieldTypes(mNotes)

	// Create a DB instance and build the base query
	db := s.db

	// apply sorting
	db = util.ApplySorting(db, sortRequest)

	// apply filtering
	db = util.ApplyFiltering(db, filterRequest)

	// apply global search
	db = util.ApplyGlobalSearch(db, searchRequest, mNotesMap)

	// Calculate the total data size without considering _size
	totalElements := db.Find(&mNotess).RowsAffected

	// Calculate the total number of pages
	totalPages := totalElements / sizeInt64
	if totalElements%sizeInt64 != 0 {
		totalPages++
	}

	// paginate
	result := db.Scopes(util.ApplyPaginate(pageInt, sizeInt)).Find(&mNotess)

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
		Content:          mNotess,
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

	util.Log("INFO", "service", "GetPageMNotes", "sort is empty: "+strconv.FormatBool(sort.Empty))

	return &page, nil
}
