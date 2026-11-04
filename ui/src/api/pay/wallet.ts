import request from '@/utils/http'

// 获取钱包账户分页列表
export function fetchGetPayWalletList(params: Api.Pay.WalletSearchParams) {
  return request.get<Api.Pay.WalletList>({
    url: '/api/v1/pay/wallets',
    params
  })
}

// 获取某会员的钱包流水
export function fetchGetPayWalletRecords(params: Api.Pay.WalletRecordSearchParams) {
  return request.get<Api.Pay.WalletRecordList>({
    url: '/api/v1/pay/wallets/records',
    params
  })
}

// 后台调整余额（单位：分，正数充值、负数扣减）
export function fetchAdjustPayWallet(data: Api.Pay.WalletAdjustParams) {
  return request.post<null>({
    url: '/api/v1/pay/wallets/adjust',
    params: data
  })
}

// 冻结余额（可用余额 → 冻结金额）
export function fetchFreezePayWallet(data: Api.Pay.WalletFreezeParams) {
  return request.post<null>({
    url: '/api/v1/pay/wallets/freeze',
    params: data
  })
}

// 解冻余额（冻结金额 → 可用余额）
export function fetchUnfreezePayWallet(data: Api.Pay.WalletFreezeParams) {
  return request.post<null>({
    url: '/api/v1/pay/wallets/unfreeze',
    params: data
  })
}
