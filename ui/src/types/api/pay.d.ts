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
  }
}
