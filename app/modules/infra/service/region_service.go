package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// SyncFromAMap 从高德全量拉取省市区并整表刷新，返回写入条数。
// 需要后台配置「Web 服务」Key（map.amap.serverKey），与前端 JS Key 不是同一种类型。
// 入参 ctx 为请求上下文；返回写入条数与业务错误码。
func (s *RegionService) SyncFromAMap(ctx context.Context) (int, error) {
	key := strings.TrimSpace(s.configs.GetString(ctx, regionServerKeyConfig, ""))
	if key == "" {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest,
			"请先在「参数配置」填写高德 Web 服务 Key（map.amap.serverKey），JS API Key 不能用于本次同步")
	}

	resp, err := fetchAmapDistricts(ctx, key)
	if err != nil {
		return 0, err
	}

	rows := flattenDistricts(resp)
	if len(rows) == 0 {
		return 0, baseErrs.WithMsg(baseErrs.BadRequest, "高德未返回行政区划数据，请检查 Key 的权限与配额")
	}

	if err := s.regions.ReplaceAll(ctx, rows); err != nil {
		return 0, err
	}

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

	var walk func(parentAdcode string, list []amapDistrict)
	walk = func(parentAdcode string, list []amapDistrict) {
		for index, item := range list {
			adcode := strings.TrimSpace(item.Adcode)
			if adcode == "" || adcode == RegionRootAdcode {
				continue
			}

			rows = append(rows, model.InfraRegionModel{
				ID:           support.NewID(),
				ParentAdcode: parentAdcode,
				Adcode:       adcode,
				Name:         item.Name,
				Level:        item.Level,
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
