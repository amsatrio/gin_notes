package model

import "github.com/amsatrio/gin_notes/model/response"

type TResetPassword struct {
	Id          uint              `form:"id" json:"id" xml:"id" gorm:"primary_key;not null;type:bigint;comment:Auto increment" binding:"required"`
	
	OldPassword string            `form:"oldPassword" json:"oldPassword" xml:"oldPassword" gorm:"size:255;type:varchar(255)" binding:"max=255"`
	NewPassword string            `form:"newPassword" json:"newPassword" xml:"newPassword" gorm:"size:255;type:varchar(255)" binding:"max=255"`
	ResetFor    string            `form:"resetFor" json:"resetFor" xml:"resetFor" gorm:"size:20;type:varchar(20)" binding:"max=20"`
	
	CreatedBy   uint              `form:"createdBy" json:"createdBy" xml:"createdBy" gorm:"not null;type:bigint"`
	CreatedOn   response.JSONTime `form:"createdOn" json:"createdOn" xml:"createdOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	ModifiedBy  uint              `form:"modifiedBy" json:"modifiedBy" xml:"modifiedBy" gorm:"type:bigint"`
	ModifiedOn  response.JSONTime `form:"modifiedOn" json:"modifiedOn" xml:"modifiedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	DeletedBy   uint              `form:"deletedBy" json:"deletedBy" xml:"deletedBy" gorm:"type:bigint"`
	DeletedOn   response.JSONTime `form:"deletedOn" json:"deletedOn" xml:"deletedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	IsDelete    *bool             `form:"isDelete" json:"isDelete" xml:"isDelete" gorm:"type:boolean;comment:default FALSE"`
}

func (TResetPassword) TableName() string {
	return "t_reset_password"
}
