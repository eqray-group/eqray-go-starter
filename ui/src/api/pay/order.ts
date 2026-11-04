import request from '@/utils/http'

// 获取支付订单分页列表
export function fetchGetPayOrderList(params: Api.Pay.OrderSearchParams) {
  return request.get<Api.Pay.OrderList>({
    url: '/api/v1/pay/orders',
    params
  })
}

// 获取支付订单详情
export function fetchGetPayOrderDetail(id: string) {
  return request.get<Api.Pay.OrderItem>({
    url: `/api/v1/pay/orders/${id}`
  })
}

// 主动向渠道查单并同步本地状态
export function fetchSyncPayOrder(id: string) {
  return request.post<Api.Pay.OrderItem>({
    url: `/api/v1/pay/orders/${id}/sync`
  })
}

// 关闭支付订单（渠道侧 + 本地）
export function fetchClosePayOrder(id: string) {
  return request.post<null>({
    url: `/api/v1/pay/orders/${id}/close`
  })
}
