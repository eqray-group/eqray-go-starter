import request from '@/utils/http'

// 采集一次系统监控数据（进程 / 内存 / DB / 缓存 / 24h 请求 / 最近错误）
export function fetchGetMonitorOverview() {
  return request.get<Api.Infra.MonitorOverview>({
    url: '/api/v1/infra/monitor/overview'
  })
}
