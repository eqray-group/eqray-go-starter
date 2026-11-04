import request from '@/utils/http'

// 获取支付渠道分页列表（config 已脱敏）
export function fetchGetPayChannelList(params: Api.Pay.ChannelSearchParams) {
  return request.get<Api.Pay.ChannelList>({
    url: '/api/v1/pay/channels',
    params
  })
}

// 按渠道编码获取动态表单字段定义
export function fetchGetPayChannelSchema(code: string) {
  return request.get<Api.Pay.ChannelSchema>({
    url: '/api/v1/pay/channels/schema',
    params: { code }
  })
}

// 按渠道编码获取系统生成的异步通知地址
export function fetchGetPayChannelNotifyUrl(code: string) {
  return request.get<{ payNotifyUrl: string }>({
    url: '/api/v1/pay/channels/notify-urls',
    params: { code }
  })
}

// 获取支付渠道详情（config 已脱敏）
export function fetchGetPayChannelDetail(id: string) {
  return request.get<Api.Pay.ChannelItem>({
    url: `/api/v1/pay/channels/${id}`
  })
}

// 新增支付渠道
export function fetchCreatePayChannel(data: Api.Pay.ChannelSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/pay/channels',
    params: data
  })
}

// 编辑支付渠道（敏感字段留空表示不修改）
export function fetchUpdatePayChannel(id: string, data: Api.Pay.ChannelSaveParams) {
  return request.put<null>({
    url: `/api/v1/pay/channels/${id}`,
    params: data
  })
}

// 启用 / 停用支付渠道
export function fetchSetPayChannelStatus(id: string, enable: boolean) {
  return request.put<null>({
    url: `/api/v1/pay/channels/${id}/status`,
    params: { enable }
  })
}

// 渠道连通性测试：用当前凭据发起一次查单
export function fetchTestPayChannel(id: string) {
  return request.post<{ ok: boolean; message: string }>({
    url: `/api/v1/pay/channels/${id}/test`
  })
}

// 删除支付渠道
export function fetchDeletePayChannel(id: string) {
  return request.del<null>({
    url: `/api/v1/pay/channels/${id}`
  })
}
