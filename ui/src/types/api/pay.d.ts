/**
 * 支付模块（渠道、订单、退款、回调日志）接口类型定义
 *
 * 全局命名空间 Api 的子命名空间，与其他模块的类型文件自动合并，业务代码无需导入即可使用：
 * `const item: Api.Pay.XxxItem = ...`
 *
 * @module types/api/pay
 */

declare namespace Api {
  /** 支付模块类型 */
  namespace Pay {
    /** 支付渠道配置：一行 = 一种支付方式 */
    interface ChannelItem {
      id: string
      /** 渠道编码，如 wx_native / ali_pc */
      code: string
      name: string
      /** 支付厂商：wx / alipay */
      channel: string
      /** 支付方式：native/h5/jsapi/mini/app/page/wap/qr */
      payMethod: string
      /** 渠道配置 JSON（已脱敏；编辑留空表示不修改） */
      config: string
      /** 0 停用 / 1 启用 */
      status: number
      sort: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface ChannelList {
      list: ChannelItem[]
      total: number
    }

    interface ChannelSearchParams {
      keyword?: string
      code?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface ChannelSaveParams {
      name: string
      /** 新增必填，编辑忽略 */
      code?: string
      /** config JSON 字符串 */
      config: string
      status: number
      sort?: number
      remark?: string
    }

    /** 渠道动态表单字段定义 */
    interface ChannelField {
      key: string
      label: string
      /** input / password / textarea / switch */
      type: string
      required: boolean
      /** 敏感字段：列表脱敏、编辑留空不改 */
      secret: boolean
      rows?: number
      desc?: string
    }

    /** 渠道 schema 响应 */
    interface ChannelSchema {
      code: string
      vendor: string
      payMethod: string
      fields: ChannelField[]
      /** 全部受支持的渠道编码 */
      codes: string[]
    }

    /** 支付订单 */
    interface OrderItem {
      id: string
      merchantOrderNo: string
      bizType: string
      bizOrderNo: string
      channelId: string
      channelCode: string
      subject: string
      description?: string
      /** 金额（分） */
      price: number
      currency: string
      /** 0 待支付 / 10 支付中 / 20 已支付 / 30 已关闭 */
      status: number
      /** 0 未退款 / 10 部分退款 / 20 全额退款 */
      refundStatus: number
      userId: string
      clientIp?: string
      channelOrderNo?: string
      channelUser?: string
      notifyUrl?: string
      expireTime?: string
      successTime?: string
      createdAt?: string
      updatedAt?: string
    }

    interface OrderList {
      list: OrderItem[]
      total: number
    }

    interface OrderSearchParams {
      merchantOrderNo?: string
      bizType?: string
      bizOrderNo?: string
      channelCode?: string
      status?: number
      /** 创建时间下限，yyyy-MM-dd HH:mm:ss */
      beginTime?: string
      /** 创建时间上限，yyyy-MM-dd HH:mm:ss */
      endTime?: string
      /** 搜索栏的日期范围（仅前端使用，提交前拆成 beginTime / endTime） */
      createTimeRange?: [string, string] | null
      pageNo?: number
      pageSize?: number
    }

    /** 退款单 */
    interface RefundItem {
      id: string
      merchantRefundNo: string
      orderId: string
      merchantOrderNo: string
      channelCode: string
      /** 退款金额（分） */
      refundPrice: number
      reason: string
      /** 0 待退款 / 10 退款中 / 20 已退款 / 30 退款失败 */
      status: number
      channelRefundNo?: string
      errorMsg?: string
      successTime?: string
      createdAt?: string
    }

    interface RefundList {
      list: RefundItem[]
      total: number
    }

    interface RefundSearchParams {
      merchantRefundNo?: string
      merchantOrderNo?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface RefundSaveParams {
      orderId: string
      /** 退款金额（分） */
      refundPrice: number
      reason: string
    }

    /** 支付回调日志 */
    interface NotifyLogItem {
      id: string
      channelCode: string
      /** pay / refund */
      type: string
      merchantOrderNo: string
      requestBody: string
      decryptedBody: string
      /** 0 失败 / 1 通过 */
      verifyStatus: number
      /** 0 失败 / 1 成功 */
      handleStatus: number
      errorMsg: string
      createdAt?: string
    }

    interface NotifyLogList {
      list: NotifyLogItem[]
      total: number
    }

    interface NotifyLogSearchParams {
      channelCode?: string
      type?: string
      merchantOrderNo?: string
      verifyStatus?: number
      handleStatus?: number
      pageNo?: number
      pageSize?: number
    }

    /** 钱包账户：一个会员一行，金额单位「分」 */
    interface WalletItem {
      id: string
      userId: string
      /** 可用余额（分） */
      balance: number
      /** 冻结金额（分），总资产 = balance + frozenBalance */
      frozenBalance: number
      /** 累计充值（分） */
      totalRecharge: number
      /** 累计消费（分） */
      totalExpense: number
      /** 0 冻结 / 1 正常 */
      status: number
      updatedAt?: string
    }

    interface WalletList {
      list: WalletItem[]
      total: number
    }

    interface WalletSearchParams {
      userId?: string
      pageNo?: number
      pageSize?: number
    }

    /** 钱包流水 */
    interface WalletRecordItem {
      id: string
      userId: string
      /** 1 充值 / 2 消费 / 3 退款 / 4 调整 */
      type: number
      /** 变动金额（分），正数入账、负数出账 */
      amount: number
      /** 变动后余额（分） */
      balanceAfter: number
      bizType: string
      bizOrderNo: string
      remark: string
      createdAt?: string
    }

    interface WalletRecordList {
      list: WalletRecordItem[]
      total: number
    }

    interface WalletRecordSearchParams {
      userId: string
      pageNo?: number
      pageSize?: number
    }

    /** 后台调账：正数充值、负数扣减（单位：分） */
    interface WalletAdjustParams {
      userId: string
      amount: number
      remark?: string
    }

    /** 后台冻结 / 解冻：金额恒为正数（单位：分） */
    interface WalletFreezeParams {
      userId: string
      amount: number
      remark?: string
    }

    /** 前端拉起支付的统一结果，kind 决定渲染方式 */
    interface ChannelPayInfo {
      /** qr 二维码 / redirect 跳转 / jsapi 小程序或公众号参数 / app 参数 / wallet 余额成功 / mock 模拟 */
      kind: string
      /** 二维码内容或跳转地址（kind=qr / redirect 时有效） */
      url?: string
      /** JSAPI / 小程序 / App 的调起参数 */
      params?: Record<string, unknown>
    }

    /** 后台联调下单结果 */
    interface TestOrderResult {
      id: string
      merchantOrderNo: string
      channelCode: string
      /** 金额（分），固定 1 */
      price: number
      /** 订单状态 */
      status: number
      payInfo: ChannelPayInfo
    }

    /** 联调下单入参 */
    interface TestOrderParams {
      userId: string
      channelCode: string
    }

    /** 批量联调：单个渠道的结果，失败时 error 有值 */
    interface TestOrderOutcome {
      channelCode: string
      id?: string
      merchantOrderNo?: string
      price: number
      status: number
      payInfo?: ChannelPayInfo
      error?: string
    }
  }
}
