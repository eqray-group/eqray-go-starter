package model

import (
	"eqray-go-starter/app/modules/base/types"
)

// InfraRegionTable 是行政区划表名常量。
const InfraRegionTable = "infra_region"

// 行政区划层级，与高德 DistrictSearch 返回的 level 一致。
const (
	RegionLevelProvince = "province"
	RegionLevelCity     = "city"
	RegionLevelDistrict = "district"
)

// InfraRegionModel 对应 infra_region 表：省 / 市 / 区县三级，用 adcode 串联。
// 数据由后台「同步行政区划」一次性拉取入库，前端级联直接查库。
type InfraRegionModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// ParentAdcode 父级区划编码，根节点为 100000
	ParentAdcode string `db:"parent_adcode" json:"parentAdcode"`
	// Adcode 区划编码（高德 adcode），与 member_address.area_id 同源
	Adcode string `db:"adcode" json:"adcode"`
	// Name 区划名称
	Name string `db:"name" json:"name"`
	// Level 层级：province / city / district
	Level string `db:"level" json:"level"`
	// Sort 同级展示顺序，按高德返回顺序
	Sort int `db:"sort" json:"sort"`
	// Status 状态：0 停用 / 1 启用
	Status int8 `db:"status" json:"status"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt types.DateTime `db:"updated_at" json:"updatedAt"`
}

// TableName 返回表名。
func (m *InfraRegionModel) TableName() string {
	return InfraRegionTable
}
