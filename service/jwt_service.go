package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/amsatrio/gin_notes/model"
	"github.com/amsatrio/gin_notes/model/request"
)

type JwtService interface {
	JwtAuthenticate(context context.Context, auth request.RequestAuth) ([]string, error)
}

type JwtServiceImpl struct {
	db *gorm.DB
}

func NewJwtServiceImpl(db *gorm.DB) JwtService {
	return &JwtServiceImpl{
		db: db,
	}
}

func (j *JwtServiceImpl) JwtAuthenticate(context context.Context, auth request.RequestAuth) ([]string, error) {
	mUser := model.MUser{}
	var roles []model.MRole
	var authorities []string

	// authenticate
	result := j.db.Preload("MRole").First(&mUser, "email = ? and password = ?", auth.Username, auth.Password)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("username / password invalid")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	// get authorities
	levelBased := false
	if levelBased {
		// get level authorities
		levelThreshold := mUser.MRole.Id
		if err := j.db.Where("level >= ?", levelThreshold).Find(&roles).Error; err != nil {
			return nil, err
		}
		for _, v := range roles {
			authorities = append(authorities, v.Code)
		}
	} else {
		// get separate authorities
		authorities = append(authorities, mUser.MRole.Code)
	}

	return authorities, nil
}
