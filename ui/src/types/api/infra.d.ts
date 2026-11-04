/**
 * 基础设施（文件、参数配置、API 日志、短信）接口类型定义
 *
 * 全局命名空间 Api 的子命名空间，与其他模块的类型文件自动合并，业务代码无需导入即可使用：
 * `const item: Api.Infra.XxxItem = ...`
 *
 * @module types/api/infra
 */

declare namespace Api {
  /** 基础设施类型 */
  namespace Infra {
    /** 短信渠道配置（对齐 yudao system_sms_channel） */
    interface SmsChannelItem {
      id: string
      name: string
      /** DebugLog / AliyunSms / TencentSms */
      code: string
      apiKey: string
      /** 短信签名 */
      signature: string
      /** 扩展配置 JSON，腾讯云为 {"region","sdkAppId"} */
      extra: string
      status: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface SmsChannelList {
      list: SmsChannelItem[]
      total: number
    }

    interface SmsChannelSearchParams {
      keyword?: string
      code?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface SmsChannelSaveParams {
      name: string
      code: string
      apiKey?: string
      /** 编辑留空表示不修改 */
      apiSecret?: string
      signature?: string
      extra?: string
      status: number
      remark?: string
    }

    /** 短信模板（对齐 yudao system_sms_template） */
    interface SmsTemplateItem {
      id: string
      channelId: string
      code: string
      name: string
      /** 1 验证码 / 2 通知 */
      type: number
      apiTemplateId: string
      content: string
      /** 模板变量名，逗号分隔 */
      params: string
      status: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface SmsTemplateList {
      list: SmsTemplateItem[]
      total: number
    }

    interface SmsTemplateSearchParams {
      keyword?: string
      channelId?: string
      type?: number
      status?: number
      pageNo?: number
      pageSize?: number
    }

    interface SmsTemplateSaveParams {
      channelId: string
      code: string
      name: string
      type: number
      apiTemplateId?: string
      content: string
      params?: string
      status: number
      remark?: string
    }

    /** 参数配置 */
    interface ConfigItem {
      id: string
      category: string
      name: string
      configKey: string
      value: string
      /** 1 系统内置 / 2 自定义 */
      type: number
      /** 0 否 / 1 是 */
      visible: number
      remark: string
      creator?: string
      createdAt?: string
      updatedAt?: string
    }

    /** 参数配置列表 */
    interface ConfigList {
      list: ConfigItem[]
      total: number
    }

    /** 参数配置搜索参数 */
    interface ConfigSearchParams {
      category?: string
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 参数配置新增参数 */
    interface ConfigCreateParams {
      category: string
      name: string
      configKey: string
      value: string
      type: number
      visible: number
      remark?: string
    }

    /** 参数配置编辑参数，参数键与类型不可修改 */
    type ConfigUpdateParams = Omit<ConfigCreateParams, 'configKey' | 'type'>

    /** 文件记录 */
    interface FileItem {
      id: string
      /** 原始文件名 */
      name: string
      /** 存储路径 */
      path: string
      /** 访问地址 */
      url: string
      /** MIME 类型 */
      type: string
      /** 文件大小（字节） */
      size: number
      createdAt: string
    }

    /** 文件列表 */
    interface FileList {
      list: FileItem[]
      total: number
    }

    /** 文件搜索参数 */
    interface FileSearchParams {
      keyword?: string
      pageNo?: number
      pageSize?: number
    }

    /** 文件配置 */
    interface FileConfigItem {
      id: string
      name: string
      /** 存储类型：local 本地磁盘 / s3 S3 兼容存储（MinIO、OSS、COS 等） */
      storage: string
      /** 1 主配置 / 0 普通 */
      master: number
      /** 存储配置 JSON */
      config: string
      createdAt?: string
    }

    /** 文件配置新增 / 编辑参数 */
    interface FileConfigParams {
      name: string
      storage: string
      master: number
      config: string
    }

    /** 文件配置搜索参数 */
    interface FileConfigSearchParams {
      name?: string
      storage?: string
      pageNo?: number
      pageSize?: number
    }

    /** API 访问日志 */
    interface ApiAccessLogItem {
      id: string
      traceId: string
      userId: string
      userType: number
      applicationName: string
      requestMethod: string
      requestUrl: string
      requestParams: string
      responseBody: string
      userIp: string
      userAgent: string
      operateModule: string
      operateName: string
      beginAt: string
      endAt: string
      duration: number
      resultCode: number
      createdAt: string
    }

    /** API 访问日志列表 */
    interface ApiAccessLogList {
      list: ApiAccessLogItem[]
      total: number
    }

    /** API 错误日志 */
    interface ApiErrorLogItem {
      id: string
      traceId: string
      userId: string
      requestMethod: string
      requestUrl: string
      requestParams: string
      userIp: string
      userAgent: string
      name: string
      message: string
      exceptionStack: string
      processStatus: number
      processTime?: string
      createdAt: string
    }

    /** API 错误日志列表 */
    interface ApiErrorLogList {
      list: ApiErrorLogItem[]
      total: number
    }

    /** API 日志搜索参数 */
    interface ApiLogSearchParams {
      requestUrl?: string
      keyword?: string
      resultCode?: number
      pageNo?: number
      pageSize?: number
    }

    /** 上传成功后的文件信息 */
    interface UploadedFile {
      /** 雪花 ID，后端序列化为字符串 */
      id: string
      /** 原始文件名 */
      name: string
      /** 可直接访问的地址 */
      url: string
      /** MIME 类型 */
      type: string
      /** 文件大小（字节） */
      size: number
    }

    /** 行政区划（省 / 市 / 区县），数据来自 infra_region 表 */
    interface RegionItem {
      /** 区划编码 adcode，如 330106 */
      adcode: string
      /** 区划名称 */
      name: string
      /** 层级：province / city / district */
      level: string
      /** 是否叶子节点（没有下级） */
      leaf: boolean
    }
  }
}
