package model

import (
	baseModel "eqray-go-starter/app/modules/base/model"
	"eqray-go-starter/app/modules/base/types"
)

const MemberSignRecordTable = "member_sign_record"

// MemberSignRecordModel 对应 member_sign_record 表，一天一条签到记录。
type MemberSignRecordModel struct {
	baseModel.BaseModel
	UserID   types.ID       `db:"user_id" json:"userId"`
	SignDate types.DateTime `db:"sign_date" json:"signDate"`
	Point    int            `db:"point" json:"point"`
}

// TableName 返回表名。
func (m *MemberSignRecordModel) TableName() string {
	return MemberSignRecordTable
}
