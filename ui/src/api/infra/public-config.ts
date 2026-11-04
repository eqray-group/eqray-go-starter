import request from '@/utils/http'

/**
 * 批量读取公开参数（免登录）
 * 只支持服务端白名单内的键（如地图 Key），未配置的键不会出现在返回结果里
 * @param keys 参数键列表，如 ['map.amap.key', 'map.amap.securityCode']
 */
export function fetchGetPublicConfigs(keys: string[]) {
  return request.get<Record<string, string>>({
    url: '/api/v1/client/infra/public-configs',
    params: { keys: keys.join(',') }
  })
}
