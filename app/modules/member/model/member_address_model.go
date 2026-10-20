package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
	"eqray-go-starter/app/modules/base/types"
)

const MemberAddressTable = "member_address"

// MemberAddressModel 对应 member_address 表，后台以只读方式展示会员收货地址。
type MemberAddressModel struct {
	baseModel.BaseModel
	UserID        types.ID `db:"user_id" json:"userId"`
	Name          string   `db:"name" json:"name"`
	Mobile        string   `db:"mobile" json:"mobile"`
	AreaID        int      `db:"area_id" json:"areaId"`
	DetailAddress string   `db:"detail_address" json:"detailAddress"`
	// DefaultStatus 0 否 / 1 是
	DefaultStatus int8 `db:"default_status" json:"defaultStatus"`
}

// TableName 返回表名。
func (m *MemberAddressModel) TableName() string {
	return MemberAddressTable
}
