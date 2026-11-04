import request from '@/utils/http'

// 获取支付回调日志分页列表
export function fetchGetPayNotifyLogList(params: Api.Pay.NotifyLogSearchParams) {
  return request.get<Api.Pay.NotifyLogList>({
    url: '/api/v1/pay/notify-logs',
    params
  })
}
