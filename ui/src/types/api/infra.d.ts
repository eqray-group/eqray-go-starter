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

    /** 后台省市县列表行：比 C 端多父级、排序与时间 */
    interface RegionRowItem {
      id: string
      /** 父级 adcode，根为 100000 */
      parentAdcode: string
      adcode: string
      name: string
      level: string
      /** 同级排序 */
      sort: number
      /** 0 停用 / 1 启用 */
      status: number
      createdAt?: string
      updatedAt?: string
    }

    /** 省市县查询参数 */
    interface RegionSearchParams {
      /** 名称或编码模糊匹配 */
      keyword?: string
      /** province / city / district */
      level?: string
      /** 父级 adcode */
      parentAdcode?: string
      pageNo?: number
      pageSize?: number
    }

    /** 区划总览统计 */
    interface RegionStats {
      total: number
      province: number
      city: number
      district: number
      /** 最后同步时间，未同步过为空串 */
      lastSyncedAt: string
    }

    /** 省市县列表返回 */
    interface RegionListResult {
      list: RegionRowItem[]
      total: number
      stats: RegionStats
    }

    /** 定时任务 */
    interface JobItem {
      id: string
      name: string
      /** 处理器标识（代码注册） */
      handler: string
      /** cron 表达式（5 段） */
      cron: string
      /** 0 停用 / 1 启用 */
      status: number
      remark: string
      nextRunAt: string | null
      lastRunAt: string | null
      /** 上次结果：0 失败 / 1 成功 */
      lastStatus: number
      /** 上次耗时（毫秒） */
      lastDuration: number
      lastOutput: string
      updatedAt?: string
    }

    interface JobList {
      list: JobItem[]
      total: number
    }

    interface JobSearchParams {
      keyword?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增任务时可选的处理器 */
    interface JobHandlerItem {
      key: string
      name: string
      description: string
      defaultCron: string
    }

    interface JobSaveParams {
      name: string
      /** 编辑时忽略 */
      handler?: string
      cron: string
      status: number
      remark?: string
    }

    /** 任务执行日志 */
    interface JobLogItem {
      id: string
      jobId: string
      jobName: string
      handler: string
      /** cron / manual */
      triggerType: string
      /** 0 失败 / 1 成功 */
      status: number
      startedAt: string
      finishedAt: string
      /** 耗时（毫秒） */
      duration: number
      output: string
      error: string
    }

    interface JobLogList {
      list: JobLogItem[]
      total: number
    }

    interface JobLogSearchParams {
      jobId?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 依赖探测结果 */
    interface MonitorHealth {
      ok: boolean
      /** 探测耗时（毫秒） */
      latencyMs: number
      /** 失败原因，正常时为空 */
      message: string
    }

    /** 最近一条接口错误日志 */
    interface MonitorErrorItem {
      id: string
      name: string
      requestUrl: string
      message: string
      createdAt: string
    }

    /** 系统监控总览 */
    interface MonitorOverview {
      app: {
        goVersion: string
        uptimeSeconds: number
        goroutines: number
        cpuCores: number
      }
      memory: {
        allocMB: number
        heapInuseMB: number
        sysMB: number
        numGC: number
        heapObjects: number
      }
      db: MonitorHealth
      cache: MonitorHealth
      request: {
        total: number
        errors: number
        slow: number
        avgDurationMs: number
      }
      recentErrors: MonitorErrorItem[]
      collectedAt: string
    }

    /** 通知渠道配置字段选项 */
    interface NoticeFieldOption {
      label: string
      value: string
    }

    /** 通知渠道配置字段定义（后端下发，前端动态渲染） */
    interface NoticeFieldSchema {
      key: string
      label: string
      /** text / password / number / select */
      type: string
      required: boolean
      placeholder?: string
      desc?: string
      options?: NoticeFieldOption[]
    }

    /** 通知渠道类型定义 */
    interface NoticeTypeSchema {
      type: string
      name: string
      desc: string
      /** 接收目标在页面上的叫法 */
      targetLabel: string
      fields: NoticeFieldSchema[]
    }

    /** 通知渠道 */
    interface NoticeChannelItem {
      id: string
      name: string
      /** email / wecom / dingtalk */
      type: string
      /** 配置 JSON 文本 */
      config: string
      status: number
      remark: string
      createdAt?: string
      updatedAt?: string
    }

    interface NoticeChannelList {
      list: NoticeChannelItem[]
      total: number
    }

    interface NoticeChannelSaveParams {
      name: string
      type: string
      config: Record<string, string>
      status: number
      remark?: string
    }

    /** 通知模板 */
    interface NoticeTemplateItem {
      id: string
      code: string
      name: string
      type: string
      title: string
      content: string
      status: number
      remark: string
      updatedAt?: string
    }

    interface NoticeTemplateList {
      list: NoticeTemplateItem[]
      total: number
    }

    interface NoticeTemplateSaveParams {
      code?: string
      name: string
      type: string
      title?: string
      content: string
      status: number
      remark?: string
    }

    /** 通知发送记录 */
    interface NoticeLogItem {
      id: string
      channelId: string
      channelName: string
      type: string
      templateCode: string
      target: string
      title: string
      content: string
      /** 0 失败 / 1 成功 */
      status: number
      error: string
      duration: number
      creator: string
      createdAt: string
    }

    interface NoticeLogList {
      list: NoticeLogItem[]
      total: number
    }

    interface NoticeSearchParams {
      keyword?: string
      type?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 通用发送入参 */
    interface NoticeSendParams {
      channelId?: string
      type?: string
      templateCode?: string
      target?: string
      title?: string
      content?: string
      variables?: Record<string, string>
    }

    /** 代码生成：可选表 */
    interface GeneratorTable {
      name: string
      comment: string
      rows: number
    }

    /** 代码生成：字段元信息 */
    interface GeneratorColumn {
      name: string
      type: string
      columnType: string
      nullable: boolean
      primaryKey: boolean
      autoIncrement: boolean
      comment: string
      goType: string
      fieldName: string
      jsonName: string
      /** input / number / textarea / switch / datetime */
      formType: string
    }

    /** 代码生成：一个输出文件 */
    interface GeneratorFile {
      path: string
      language: string
      /** false 表示仅预览（如路由片段） */
      writable: boolean
      content: string
    }

    interface GeneratorParams {
      table: string
      module: string
      pascal?: string
      overwrite?: boolean
    }

    interface GeneratorResult {
      written: string[]
      skipped: string[]
    }
  }
}
