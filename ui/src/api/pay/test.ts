import request from '@/utils/http'

// 创建 1 分钱联调测试单并发起支付
export function fetchCreatePayTestOrder(data: Api.Pay.TestOrderParams) {
  return request.post<Api.Pay.TestOrderResult>({
    url: '/api/v1/pay/test-orders',
    params: data
  })
}

// 批量创建联调测试单：channelCodes 为空表示全部启用渠道
export function fetchCreatePayTestOrdersBatch(data: { userId: string; channelCodes?: string[] }) {
  return request.post<Api.Pay.TestOrderOutcome[]>({
    url: '/api/v1/pay/test-orders/batch',
    params: data
  })
}

// 查询联调单当前状态
export function fetchGetPayTestOrder(merchantOrderNo: string) {
  return request.get<Api.Pay.OrderItem>({
    url: `/api/v1/pay/test-orders/${merchantOrderNo}`
  })
}

// 主动向渠道查单并同步（用户已付款但回调未达时使用）
export function fetchSyncPayTestOrder(merchantOrderNo: string) {
  return request.post<Api.Pay.OrderItem>({
    url: `/api/v1/pay/test-orders/${merchantOrderNo}/sync`
  })
}

// 模拟渠道：把订单置为已支付
export function fetchMockPayTestOrder(merchantOrderNo: string) {
  return request.post<Api.Pay.OrderItem>({
    url: `/api/v1/pay/test-orders/${merchantOrderNo}/mock-pay`
  })
}
