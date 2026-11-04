import request from '@/utils/http'

/**
 * 查询某级区划的下级列表（省 / 市 / 区县）
 * 数据已落库，接口只读且免登录，可在前端长期缓存
 * @param parentAdcode 父级 adcode，空表示查省份（根节点 100000）
 */
export function fetchGetRegionChildren(parentAdcode?: string) {
  return request.get<Api.Infra.RegionItem[]>({
    url: '/api/v1/client/infra/regions',
    params: { parentAdcode }
  })
}

/**
 * 后台查询省市县列表（含各层级条数与最后同步时间）
 * 数据已落库，只读分页查询
 */
export function fetchGetInfraRegionList(params: Api.Infra.RegionSearchParams) {
  return request.get<Api.Infra.RegionListResult>({
    url: '/api/v1/infra/regions',
    params
  })
}

/**
 * 后台触发一次全量同步（从高德拉取省市区并整表刷新）
 * 需要先在参数配置中填写 map.amap.serverKey（Web 服务 Key）
 */
export function fetchSyncInfraRegions() {
  return request.post<{ count: number }>({
    url: '/api/v1/infra/regions/sync',
    // 全量同步要请求高德 + 重刷 3000+ 条，默认 15s 不够，单独放宽到 3 分钟
    timeout: 180000
  })
}
