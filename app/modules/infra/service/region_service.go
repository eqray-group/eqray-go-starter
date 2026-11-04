package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"eqray-go-starter/app/facades"
	baseErrs "eqray-go-starter/app/modules/base/errs"
	"eqray-go-starter/app/modules/base/support"
	"eqray-go-starter/app/modules/base/types"
	"eqray-go-starter/app/modules/infra/model"
	"eqray-go-starter/app/modules/infra/repository"
)

const (
	// RegionRootAdcode 是全国根节点，前端级联第一层用它查省
	RegionRootAdcode = repository.RegionRootAdcode
	// regionServerKeyConfig 是后端拉取行政区划用的「Web 服务」Key 参数名（不对外公开）
	regionServerKeyConfig = "map.amap.serverKey"
	// amapDistrictURL 是高德行政区划查询接口
	amapDistrictURL = "https://restapi.amap.com/v3/config/district"
	// amapSubdistrict 返回层级：省、市、区县共 3 级
	amapSubdistrict = 3
)

// RegionItem 是级联选择需要的一条区划数据。
type RegionItem struct {
	// Adcode 区划编码
	Adcode string `json:"adcode"`
	// Name 区划名称
	Name string `json:"name"`
	// Level 层级：province / city / district
	Level string `json:"level"`
	// Leaf 是否叶子节点（区县为 true，直辖市下的区县也是 true）
	Leaf bool `json:"leaf"`
}

// RegionListQuery 是后台区划列表的查询条件。
type RegionListQuery struct {
	// Keyword 名称或编码模糊匹配
	Keyword string
	// Level 层级：province / city / district，空表示全部
	Level string
	// ParentAdcode 父级编码，空表示全部
	ParentAdcode string
	// PageNo / PageSize 分页参数
	PageNo   int
	PageSize int
}

// RegionStats 是区划总览：各层级条数与最后同步时间。
type RegionStats struct {
	// Total 总条数
	Total int64 `json:"total"`
	// Province 省级条数
	Province int64 `json:"province"`
	// City 市级条数
	City int64 `json:"city"`
	// District 区县条数
	District int64 `json:"district"`
	// LastSyncedAt 最后同步时间，未同步过为空串
	LastSyncedAt string `json:"lastSyncedAt"`
}

// RegionPage 是后台区划列表的返回结构。
type RegionPage struct {
	List  []model.InfraRegionModel `json:"list"`
	Total int64                    `json:"total"`
	Stats RegionStats              `json:"stats"`
}

// RegionService 负责行政区划的查询与一次性同步。
// 数据落 infra_region 表，C 端级联只查库，不再实时调高德。
type RegionService struct {
	regions *repository.RegionRepository
	configs *ConfigService
}

// NewRegionService 构造行政区划服务。
func NewRegionService() *RegionService {
	return &RegionService{
		regions: repository.NewRegionRepository(),
		configs: NewConfigService(),
	}
}

// Children 查询某级区划的下级列表，供前端级联懒加载。
// 入参 parentAdcode 为父级编码，空串表示查询省份（根为 100000）；
// 返回下级区划列表，未初始化时返回业务错误码提示先同步。
func (s *RegionService) Children(ctx context.Context, parentAdcode string) ([]RegionItem, error) {
	parentAdcode = strings.TrimSpace(parentAdcode)
	if parentAdcode == "" {
		parentAdcode = RegionRootAdcode
	}

	list, err := s.regions.ListChildren(ctx, parentAdcode)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		if parentAdcode == RegionRootAdcode {
			return nil, baseErrs.WithMsg(baseErrs.BadRequest, "行政区划数据未初始化，请先在后台执行「同步行政区划」")
		}

		return make([]RegionItem, 0), nil
	}

	parents := make([]string, 0, len(list))
	for _, item := range list {
		parents = append(parents, item.Adcode)
	}

	counts, err := s.regions.CountByParents(ctx, parents)
	if err != nil {
		return nil, err
	}

	items := make([]RegionItem, 0, len(list))
	for _, item := range list {
		items = append(items, RegionItem{
			Adcode: item.Adcode,
			Name:   item.Name,
			Level:  item.Level,
			Leaf:   counts[item.Adcode] == 0,
		})
	}

	return items, nil
}

// Page 分页查询区划，供后台「省市县」页查看，并附带各层级条数与最后同步时间。
// 入参 q 为查询条件；返回列表、总数与总览统计。
func (s *RegionService) Page(ctx context.Context, q RegionListQuery) (*RegionPage, error) {
	pageNo, pageSize := q.PageNo, q.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	list, total, err := s.regions.Paginate(ctx, repository.RegionFilter{
		Keyword:      q.Keyword,
		Level:        q.Level,
		ParentAdcode: strings.TrimSpace(q.ParentAdcode),
	}, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	counts, err := s.regions.CountByLevel(ctx)
	if err != nil {
		return nil, err
	}
	lastSyncedAt, err := s.regions.LastSyncedAt(ctx)
	if err != nil {
		return nil, err
	}

	var statsTotal int64
	for _, count := range counts {
		statsTotal += count
	}

	return &RegionPage{
		List:  list,
		Total: total,
		Stats: RegionStats{
			Total:        statsTotal,
			Province:     counts["province"],
			City:         counts["city"],
			District:     counts["district"],
			LastSyncedAt: lastSyncedAt,
		},
	}, nil
}

// regionSyncMu 保证同一时间只有一个同步任务，避免两次同步交叉删表插入导致锁等待。
var regionSyncMu sync.Mutex

// SyncFromAMap 从高德全量拉取省市区并整表刷新，返回写入条数。
// 需要后台配置「Web 服务」Key（map.amap.serverKey），与前端 JS Key 不是同一种类型。
// 入参 ctx 为请求上下文；返回写入条数与业务错误码。
func (s *RegionService) SyncFromAMap(ctx context.Context) (int, error) {
	if !regionSyncMu.TryLock() {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "已有同步任务在执行，请稍后再试")
	}
	defer regionSyncMu.Unlock()

	key := strings.TrimSpace(s.configs.GetString(ctx, regionServerKeyConfig, ""))
	if key == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest,
			"请先在「参数配置」填写高德 Web 服务 Key（map.amap.serverKey），JS API Key 不能用于本次同步")
	}

	startedAt := time.Now()
	resp, err := fetchAmapDistricts(ctx, key)
	if err != nil {
		facades.Log().Errorf("[行政区划同步] 拉取高德数据失败：%v", err)
		return 0, err
	}

	rows := flattenDistricts(resp)
	if len(rows) == 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest,
			"高德未返回行政区划数据，请确认该 Key 已开通「行政区划查询」且配额未用尽")
	}

	if err := s.regions.ReplaceAll(ctx, rows); err != nil {
		// 写库错误原样带出，后台页面能直接看到真实原因（表不存在 / 字段超长 / 唯一键冲突等）
		facades.Log().Errorf("[行政区划同步] 写库失败（%d 条）：%v", len(rows), err)
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "写入行政区划失败："+err.Error())
	}

	facades.Log().Infof("[行政区划同步] 完成：%d 条，耗时 %d ms", len(rows), time.Since(startedAt).Milliseconds())

	return len(rows), nil
}

// amapDistrictResponse 是高德行政区划接口的响应体。
type amapDistrictResponse struct {
	Status    string         `json:"status"`
	Info      string         `json:"info"`
	Districts []amapDistrict `json:"districts"`
}

// amapDistrict 是高德返回的单个区划节点。
type amapDistrict struct {
	Adcode    string         `json:"adcode"`
	Name      string         `json:"name"`
	Level     string         `json:"level"`
	Districts []amapDistrict `json:"districts"`
}

// fetchAmapDistricts 调高德行政区划接口，一次拿全国省市区三级。
// 入参 key 为 Web 服务 Key；返回解析后的响应。
func fetchAmapDistricts(ctx context.Context, key string) (*amapDistrictResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s?key=%s&keywords=%s&subdistrict=%d&extensions=base",
		amapDistrictURL, key, RegionRootAdcode, amapSubdistrict)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "构造高德请求失败")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "调用高德接口失败："+err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, fmt.Sprintf("高德接口返回状态码 %d", resp.StatusCode))
	}

	var body amapDistrictResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "解析高德返回失败")
	}
	if body.Status != "1" {
		return nil, baseErrs.WithMsg(baseErrs.BadRequest, "高德接口返回错误："+body.Info)
	}

	return &body, nil
}

// flattenDistricts 把高德的三级树拍平成可入库的行。
// 入参 resp 为高德响应；返回待写入的行，父级统一用 adcode 关联。
func flattenDistricts(resp *amapDistrictResponse) []model.InfraRegionModel {
	rows := make([]model.InfraRegionModel, 0, 4000)
	now := types.Now()
	// 高德偶发返回重复 adcode（如省直辖节点），按首次出现保留，避免撞唯一键
	seen := make(map[string]struct{}, 4000)

	var walk func(parentAdcode string, list []amapDistrict)
	walk = func(parentAdcode string, list []amapDistrict) {
		for index, item := range list {
			adcode := strings.TrimSpace(item.Adcode)
			if adcode == "" || adcode == RegionRootAdcode {
				continue
			}
			if _, exists := seen[adcode]; exists {
				continue
			}
			seen[adcode] = struct{}{}

			rows = append(rows, model.InfraRegionModel{
				ID:           support.NewID(),
				ParentAdcode: parentAdcode,
				Adcode:       adcode,
				Name:         truncateRunes(item.Name, 64),
				Level:        truncateRunes(item.Level, 16),
				Sort:         index,
				Status:       1,
				CreatedAt:    now,
				UpdatedAt:    now,
			})

			if len(item.Districts) > 0 {
				walk(adcode, item.Districts)
			}
		}
	}

	if len(resp.Districts) > 0 {
		// 响应第一层是「中国」节点，它的 districts 才是省
		walk(RegionRootAdcode, resp.Districts[0].Districts)
	}

	return rows
}

// truncateRunes 按字符数截断，避免超出表字段长度导致整批写入失败。
// 入参 value 为原始值、max 为最大字符数；返回截断后的字符串。
func truncateRunes(value string, max int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}

	return string(runes[:max])
}
