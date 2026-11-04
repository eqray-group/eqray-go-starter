import request from '@/utils/http'

// 获取退款单分页列表
export function fetchGetPayRefundList(params: Api.Pay.RefundSearchParams) {
  return request.get<Api.Pay.RefundList>({
    url: '/api/v1/pay/refunds',
    params
  })
}

// 获取退款单详情
export function fetchGetPayRefundDetail(id: string) {
  return request.get<Api.Pay.RefundItem>({
    url: `/api/v1/pay/refunds/${id}`
  })
}

// 发起退款：无人工审核，提交后直接调渠道
export function fetchCreatePayRefund(data: Api.Pay.RefundSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/pay/refunds',
    params: data
  })
}
