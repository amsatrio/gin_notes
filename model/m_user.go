package model

import "github.com/amsatrio/gin_notes/model/response"

type MUser struct {
	Id           uint              `form:"id" json:"id" xml:"id" gorm:"primary_key;not null;type:bigint;comment:Auto increment" binding:"required"`
	BiodataId    uint              `form:"biodataId" json:"biodataId" xml:"biodataId" gorm:"type:bigint"`
	RoleId       uint              `form:"roleId" json:"roleId" xml:"roleId" gorm:"type:bigint"`
	Email        string            `form:"email" json:"email" xml:"email" gorm:"size:100;type:varchar(100)" binding:"max=100"`
	Password     string            `form:"password" json:"password" xml:"password" gorm:"size:255;type:varchar(255)" binding:"max=255"`
	LoginAttempt int               `form:"loginAttempt" json:"loginAttempt" xml:"loginAttempt" gorm:"type:int"`
	IsLocked     bool              `form:"isLocked" json:"isLocked" xml:"isLocked" gorm:"type:boolean"`
	LastLogin    response.JSONTime `form:"lastLogin" json:"lastLogin" xml:"lastLogin" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	CreatedBy    uint              `form:"createdBy" json:"createdBy" xml:"createdBy" gorm:"not null;type:bigint"`
	CreatedOn    response.JSONTime `form:"createdOn" json:"createdOn" xml:"createdOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	ModifiedBy   uint              `form:"modifiedBy" json:"modifiedBy" xml:"modifiedBy" gorm:"type:bigint"`
	ModifiedOn   response.JSONTime `form:"modifiedOn" json:"modifiedOn" xml:"modifiedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	DeletedBy    uint              `form:"deletedBy" json:"deletedBy" xml:"deletedBy" gorm:"type:bigint"`
	DeletedOn    response.JSONTime `form:"deletedOn" json:"deletedOn" xml:"deletedOn" gorm:"type:datetime" swaggertype:"string" example:"2024-02-16 10:33:10"`
	IsDelete     *bool             `form:"isDelete" json:"isDelete" xml:"isDelete" gorm:"type:boolean;comment:default FALSE"`

	MBiodata MBiodata `gorm:"foreignKey:BiodataId"`
	MRole    MRole    `gorm:"foreignKey:RoleId"`
}

func (MUser) TableName() string {
	return "m_user"
}
