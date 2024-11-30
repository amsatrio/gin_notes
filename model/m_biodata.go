package model

import "github.com/amsatrio/gin_notes/model/response"

type MBiodata struct {
	Id          uint              `form:"id" json:"id" xml:"id" gorm:"primary_key;not null;type:bigint;comment:Auto increment" binding:"required"`
	Fullname    string            `form:"fullname" json:"fullname" xml:"fullname" gorm:"size:255;type:varchar(255)" binding:"max=255"`
	MobilePhone string            `form:"mobilePhone" json:"mobilePhone" xml:"mobilePhone" gorm:"size:15;type:varchar(15)" binding:"max=15"`
	Image       []byte            `form:"image" json:"image" xml:"image" gorm:"type:blob"`
	ImagePath   string            `form:"imagePath" json:"imagePath" xml:"imagePath" gorm:"size:255;type:varchar(255)" binding:"max=255"`
	CreatedBy   uint              `form:"createdBy" json:"createdBy" xml:"createdBy" gorm:"not null;type:bigint"`
	CreatedOn   response.JSONTime `form:"createdOn" json:"createdOn" xml:"createdOn" gorm:"not null;type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	ModifiedBy  uint              `form:"modifiedBy" json:"modifiedBy" xml:"modifiedBy" gorm:"type:bigint"`
	ModifiedOn  response.JSONTime `form:"modifiedOn" json:"modifiedOn" xml:"modifiedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	DeletedBy   uint              `form:"deletedBy" json:"deletedBy" xml:"deletedBy" gorm:"type:bigint"`
	DeletedOn   response.JSONTime `form:"deletedOn" json:"deletedOn" xml:"deletedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	IsDelete    *bool             `form:"isDelete" json:"isDelete" xml:"isDelete" gorm:"not null;type:boolean;comment:default FALSE"`
}

func (MBiodata) TableName() string {
	return "m_biodata"
}
