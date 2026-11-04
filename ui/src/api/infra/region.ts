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
 * 后台触发一次全量同步（从高德拉取省市区并整表刷新）
 * 需要先在参数配置中填写 map.amap.serverKey（Web 服务 Key）
 */
export function fetchSyncInfraRegions() {
  return request.post<{ count: number }>({
    url: '/api/v1/infra/regions/sync'
  })
}
