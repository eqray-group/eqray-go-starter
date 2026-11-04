package migrations

import (
	"eqray-go-starter/app/facades"
)

// M20261104000002CreateInfraRegionTable 建 infra_region 行政区划表。
// 省 / 市 / 区县三级一次性落库，前端级联直接查库，不再每次调用高德接口。
type M20261104000002CreateInfraRegionTable struct{}

// Signature The unique signature for the migration.
func (r *M20261104000002CreateInfraRegionTable) Signature() string {
	return "20261104000002_create_infra_region_table"
}

// Up 建 infra_region 表。
// 用 adcode 串联上下级（parent_adcode），行政区划调整时整表重刷即可，不依赖自增 ID 稳定。
func (r *M20261104000002CreateInfraRegionTable) Up() error {
	if facades.Schema().HasTable("infra_region") {
		return nil
	}

	return facades.DB().Statement(`
CREATE TABLE infra_region (
  id            bigint      NOT NULL COMMENT '主键',
  parent_adcode varchar(12) NOT NULL DEFAULT '' COMMENT '父级区划编码，根节点为 100000',
  adcode        varchar(12) NOT NULL COMMENT '区划编码（高德 adcode）',
  name          varchar(64) NOT NULL COMMENT '区划名称',
  level         varchar(16) NOT NULL COMMENT '层级：province / city / district',
  sort          int         NOT NULL DEFAULT 0 COMMENT '同级展示顺序',
  status        tinyint     NOT NULL DEFAULT 1 COMMENT '状态：0 停用 / 1 启用',
  created_at    datetime    NULL COMMENT '创建时间',
  updated_at    datetime    NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_adcode (adcode),
  KEY idx_parent (parent_adcode, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行政区划表（省 / 市 / 区县）'`)
}

// Down 删除行政区划表。
func (r *M20261104000002CreateInfraRegionTable) Down() error {
	return facades.Schema().DropIfExists("infra_region")
}
