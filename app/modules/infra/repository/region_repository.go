package repository

import (
	"context"
	"strings"

	"github.com/goravel/framework/contracts/database/db"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/infra/model"
)

const regionTable = model.InfraRegionTable

// RegionRootAdcode 是全国根节点的 adcode，前端级联第一层用它查省。
const RegionRootAdcode = "100000"

// regionBatchSize 是整表刷新时的单批插入条数：1000 行 9 列约 9000 个占位符，远低于 MySQL 上限，
// 又能把 3000+ 条的重刷控制在 3~4 条 SQL 内，缩短同步耗时。
const regionBatchSize = 1000

// RegionFilter 是后台区划列表的筛选条件。
type RegionFilter struct {
	// Keyword 名称或编码模糊匹配，空表示不限
	Keyword string
	// Level 层级精确匹配（province / city / district），空表示全部
	Level string
	// ParentAdcode 父级编码精确匹配，空表示全部
	ParentAdcode string
}

// RegionRepository 是 infra_region 表的仓储。
type RegionRepository struct{}

// applyFilter 把筛选条件套到查询上；列表与计数复用同一份条件，避免两边不一致。
// 入参 query 为查询构造器、filter 为筛选条件；返回加了条件的查询。
func (r *RegionRepository) applyFilter(query db.Query, filter RegionFilter) db.Query {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		// 名称与编码拼一起匹配，避免 OR 条件和其他 AND 条件混在一起需要括号
		query = query.Where("CONCAT(name, ' ', adcode) LIKE ?", "%"+keyword+"%")
	}
	if filter.Level != "" {
		query = query.Where("level", filter.Level)
	}
	if filter.ParentAdcode != "" {
		query = query.Where("parent_adcode", filter.ParentAdcode)
	}

	return query
}

// Paginate 分页查询区划，按 adcode 升序（同一父级下顺序基本等于行政区划顺序）。
// 入参 filter 为筛选条件、pageNo/pageSize 为分页参数；返回列表与总数。
func (r *RegionRepository) Paginate(ctx context.Context, filter RegionFilter, pageNo, pageSize int) ([]model.InfraRegionModel, int64, error) {
	total, err := r.applyFilter(facades.DB().WithContext(ctx).Table(regionTable), filter).Count()
	if err != nil {
		return nil, 0, err
	}

	list := make([]model.InfraRegionModel, 0)
	if total == 0 {
		return list, 0, nil
	}

	if err := r.applyFilter(facades.DB().WithContext(ctx).Table(regionTable), filter).
		OrderBy("adcode").
		Offset(uint64((pageNo - 1) * pageSize)).
		Limit(uint64(pageSize)).
		Get(&list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CountByLevel 按层级统计条数，用于后台展示省 / 市 / 区县分布。
// 入参 ctx 为请求上下文；返回「层级 → 条数」。
func (r *RegionRepository) CountByLevel(ctx context.Context) (map[string]int64, error) {
	rows := make([]struct {
		Level string `db:"level"`
		Total int64  `db:"total"`
	}, 0)

	if err := facades.DB().WithContext(ctx).Table(regionTable).
		Select("level, COUNT(*) AS total").
		GroupBy("level").
		Get(&rows); err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Level] = row.Total
	}

	return counts, nil
}

// LastSyncedAt 返回最后一次同步时间（取全表 updated_at 最大值），未同步过返回空串。
// 入参 ctx 为请求上下文；返回格式化时间与查询错误。
func (r *RegionRepository) LastSyncedAt(ctx context.Context) (string, error) {
	row := struct {
		LastSynced string `db:"last_synced"`
	}{}

	if err := facades.DB().WithContext(ctx).Table(regionTable).
		Select("IFNULL(DATE_FORMAT(MAX(updated_at), '%Y-%m-%d %H:%i:%s'), '') AS last_synced").
		First(&row); err != nil {
		return "", err
	}

	return row.LastSynced, nil
}

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
		// 带主键条件删除：既清空全表，又能兼容开启 sql_safe_updates 的实例（无 where 的 delete 会被拒）
		if _, err := tx.Table(regionTable).Where("id", ">", 0).Delete(); err != nil {
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
