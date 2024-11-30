package model

import "github.com/amsatrio/gin_notes/model/response"

type TToken struct {
	Id uint `form:"id" json:"id" xml:"id" gorm:"primary_key;not null;type:bigint;comment:Auto increment" binding:"required"`

	Email     string            `form:"email" json:"email" xml:"email" gorm:"size:100;type:varchar(100)" binding:"max=100"`
	UserId    uint              `form:"userId" json:"userId" xml:"userId" gorm:"type:bigint"`
	Token     string            `form:"token" json:"token" xml:"token" gorm:"size:50;type:varchar(50)" binding:"max=50"`
	ExpiredOn response.JSONTime `form:"expiredOn" json:"expiredOn" xml:"expiredOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	IsExpired bool              `form:"isExpired" json:"isExpired" xml:"isExpired" gorm:"type:boolean"`
	UsedFor   string            `form:"usedFor" json:"usedFor" xml:"usedFor" gorm:"size:20;type:varchar(20)" binding:"max=20"`

	CreatedBy  uint              `form:"createdBy" json:"createdBy" xml:"createdBy" gorm:"not null;type:bigint"`
	CreatedOn  response.JSONTime `form:"createdOn" json:"createdOn" xml:"createdOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	ModifiedBy uint              `form:"modifiedBy" json:"modifiedBy" xml:"modifiedBy" gorm:"type:bigint"`
	ModifiedOn response.JSONTime `form:"modifiedOn" json:"modifiedOn" xml:"modifiedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	DeletedBy  uint              `form:"deletedBy" json:"deletedBy" xml:"deletedBy" gorm:"type:bigint"`
	DeletedOn  response.JSONTime `form:"deletedOn" json:"deletedOn" xml:"deletedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	IsDelete   *bool             `form:"isDelete" json:"isDelete" xml:"isDelete" gorm:"type:boolean;comment:default FALSE"`

	MUser MUser `gorm:"foreignKey:UserId"`
}

func (TToken) TableName() string {
	return "t_token"
}
