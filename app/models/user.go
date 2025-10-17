package models

import "github.com/goravel/framework/database/orm"

type MasterUser struct {
	orm.Model
	ID        int    `gorm:"column:id;primaryKey;autoIncrement"`
	Nama      string `gorm:"column:nama"`
	Username  string `gorm:"column:username"`
	Password  string `gorm:"column:password"`
	Role      string `gorm:"column:role"`
	KodeSp    string `gorm:"column:kode_sp"`
	IsDeleted bool   `gorm:"column:is_deleted"`
}

func (MasterUser) TableName() string {
	return "master_user"
}
