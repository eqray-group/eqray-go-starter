package repository

import (
	"context"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/infra/model"
)

const regionTable = model.InfraRegionTable

// RegionRootAdcode 是全国根节点的 adcode，前端级联第一层用它查省。
const RegionRootAdcode = "100000"

// regionBatchSize 是整表刷新时的单批插入条数，避免拼出超长 SQL。
const regionBatchSize = 500

// RegionRepository 是 infra_region 表的仓储。
type RegionRepository struct{}

// NewRegionRepository 构造行政区划仓储。
func NewRegionRepository() *RegionRepository {
	return &RegionRepository{}
}

// ListChildren 查询某父级下启用中的区划，按 sort、adcode 升序。
// 入参 parentAdcode 为父级编码；返回区划列表。
func (r *RegionRepository) ListChildren(ctx context.Context, parentAdcode string) ([]model.InfraRegionModel, error) {
	list := make([]model.InfraRegionModel, 0)

	if err := facades.DB().WithContext(ctx).Table(regionTable).
		Where("parent_adcode", parentAdcode).
		Where("status", 1).
		OrderBy("sort").
		OrderBy("adcode").
		Get(&list); err != nil {
		return nil, err
	}

	return list, nil
}

// CountByParents 统计每个父级下的子区划数量，用于判断级联节点是否为叶子。
// 入参 parents 为父级编码列表；返回「父级编码 → 子数量」。
func (r *RegionRepository) CountByParents(ctx context.Context, parents []string) (map[string]int64, error) {
	counts := make(map[string]int64, len(parents))
	if len(parents) == 0 {
		return counts, nil
	}

	rows := make([]struct {
		ParentAdcode string `db:"parent_adcode"`
		Total        int64  `db:"total"`
	}, 0, len(parents))

	args := make([]any, 0, len(parents))
	for _, parent := range parents {
		args = append(args, parent)
	}

	if err := facades.DB().WithContext(ctx).Table(regionTable).
		Select("parent_adcode, COUNT(*) AS total").
		WhereIn("parent_adcode", args).
		Where("status", 1).
		GroupBy("parent_adcode").
		Get(&rows); err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.ParentAdcode] = row.Total
	}

	return counts, nil
}

// Count 返回区划总条数，用于判断是否已初始化。
// 入参 ctx 为请求上下文；返回总条数与查询错误。
func (r *RegionRepository) Count(ctx context.Context) (int64, error) {
	return facades.DB().WithContext(ctx).Table(regionTable).Count()
}

// ReplaceAll 在事务内整表替换：先清空再分批写入。
// 行政区划每年都有调整，整表重刷比逐条 diff 简单可靠，也能清掉已撤销的区划。
// 入参 rows 为待写入的全量数据；返回写库错误。
func (r *RegionRepository) ReplaceAll(ctx context.Context, rows []model.InfraRegionModel) error {
	return facades.DB().WithContext(ctx).Transaction(func(tx db.Tx) error {
		if _, err := tx.Table(regionTable).Delete(); err != nil {
			return err
		}

		for start := 0; start < len(rows); start += regionBatchSize {
			end := start + regionBatchSize
			if end > len(rows) {
				end = len(rows)
			}

			batch := make([]map[string]any, 0, end-start)
			for _, row := range rows[start:end] {
				batch = append(batch, map[string]any{
					"id":            row.ID,
					"parent_adcode": row.ParentAdcode,
					"adcode":        row.Adcode,
					"name":          row.Name,
					"level":         row.Level,
					"sort":          row.Sort,
					"status":        row.Status,
					"created_at":    row.CreatedAt,
					"updated_at":    row.UpdatedAt,
				})
			}

			if _, err := tx.Table(regionTable).Insert(batch); err != nil {
				return err
			}
		}

		return nil
	})
}
