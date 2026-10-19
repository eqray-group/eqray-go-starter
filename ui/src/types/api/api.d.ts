/**
 * API 接口类型定义模块
 *
 * 提供所有后端接口的类型定义
 *
 * ## 主要功能
 *
 * - 通用类型（分页参数、响应结构等）
 * - 认证类型（登录、用户信息等）
 * - 系统管理类型（用户、角色等）
 * - 全局命名空间声明
 *
 * ## 使用场景
 *
 * - API 请求参数类型约束
 * - API 响应数据类型定义
 * - 接口文档类型同步
 *
 * ## 注意事项
 *
 * - 在 .vue 文件使用需要在 eslint.config.mjs 中配置 globals: { Api: 'readonly' }
 * - 使用全局命名空间，无需导入即可使用
 *
 * ## 使用方式
 *
 * ```typescript
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const response: Api.Auth.UserInfo = await fetchUserInfo()
 * ```
 *
 * @module types/api/api
 * @author Art Design Pro Team
 */

declare namespace Api {
  /** 通用类型 */
  namespace Common {
    /** 分页参数 */
    interface PaginationParams {
      /** 当前页码 */
      current: number
      /** 每页条数 */
      size: number
      /** 总条数 */
      total: number
    }

    /** 通用搜索参数 */
    type CommonSearchParams = Pick<PaginationParams, 'current' | 'size'>

    /** 分页响应基础结构 */
    interface PaginatedResponse<T = any> {
      records: T[]
      current: number
      size: number
      total: number
    }

    /** 启用状态 */
    type EnableStatus = '1' | '2'
  }

  /** 认证类型 */
  namespace Auth {
    /** 登录参数（与后端 auth/login 对齐） */
    interface LoginParams {
      username: string
      password: string
      captchaId: string
      captchaCode: string
    }

    /** 登录 / 续期响应：访问令牌短期有效，刷新令牌用于换新 */
    interface LoginResponse {
      accessToken: string
      refreshToken: string
      /** 访问令牌有效期（秒） */
      expiresIn: number
      /** 刷新令牌有效期（秒） */
      refreshExpiresIn: number
    }

    /** 图片验证码 */
    interface Captcha {
      captchaId: string
      image: string
    }

    /** 用户信息 */
    interface UserInfo {
      buttons: string[]
      roles: string[]
      /** 雪花 ID，后端序列化为字符串 */
      userId: string
      userName: string
      nickname?: string
      email: string
      mobile?: string | null
      /** 0 未知 / 1 男 / 2 女 */
      sex?: number
      avatar?: string
    }

    /** 个人资料更新参数 */
    interface ProfileUpdateParams {
      nickname: string
      email?: string
      mobile?: string
      sex: number
      avatar?: string
    }

    /** 修改密码参数 */
    interface ProfilePasswordParams {
      oldPassword: string
      newPassword: string
    }
  }

  /** 基础设施类型 */
  namespace Infra {
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
  }

  /** 控制台统计 */
  namespace Dashboard {
    /** 趋势数据点 */
    interface TrendItem {
      /** 月-日，如 10-19 */
      date: string
      count: number
    }

    /** 最近登录记录 */
    interface LoginItem {
      username: string
      userIp: string
      /** 0 失败 / 1 成功 */
      result: number
      createdAt: string
    }

    /** 控制台聚合数据 */
    interface Overview {
      counts: {
        user: number
        role: number
        menu: number
        file: number
        fileSize: number
        todayLogin: number
        todayRequest: number
        todayError: number
      }
      loginTrend: TrendItem[]
      requestTrend: TrendItem[]
      recentLogins: LoginItem[]
      system: {
        goVersion: string
        serverTime: string
      }
    }
  }

  /** 系统管理类型 */
  namespace SystemManage {
    /** 用户列表（与后端 Base.Page 对齐） */
    interface UserList {
      list: UserListItem[]
      total: number
    }

    /** 用户列表项（与后端 model.SystemUser 对齐，ID 统一为字符串） */
    interface UserListItem {
      id: string
      username: string
      nickname: string
      remark: string
      email: string | null
      mobile: string | null
      /** 0 未知 / 1 男 / 2 女 */
      sex: number
      avatar: string
      /** 0 禁用 / 1 启用 */
      status: number
      loginIp: string
      lastLoginAt: string | null
      creator: string
      createdAt: string
      updater: string
      updatedAt: string
    }

    /** 用户搜索参数（后端只支持这些筛选条件） */
    interface UserSearchParams {
      username?: string
      nickname?: string
      mobile?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑用户参数 */
    interface UserSaveParams {
      username: string
      /** 仅新增时必填，编辑不改密码 */
      password?: string
      nickname?: string
      remark?: string
      email?: string
      mobile?: string
      sex?: number
      avatar?: string
      status: number
    }

    /** 重置密码参数 */
    interface UserResetPasswordParams {
      password: string
    }

    /** 菜单树节点（与后端 model.SystemMenuModel 对齐） */
    interface MenuItem {
      id: string
      name: string
      /** 权限标识，仅按钮类填写 */
      permission: string
      /** 1 目录 / 2 菜单 / 3 按钮 */
      type: number
      sort: number
      parentId: string
      path: string
      icon: string
      component: string
      componentName: string
      /** 0 禁用 / 1 启用 */
      status: number
      /** 0 隐藏 / 1 显示 */
      visible: number
      keepAlive: number
      alwaysShow: number
      createdAt: string
      updatedAt: string
      children: MenuItem[]
    }

    /** 角色列表（与后端 Base.Page 对齐） */
    interface RoleList {
      list: RoleItem[]
      total: number
    }

    /** 角色项 */
    interface RoleItem {
      id: string
      name: string
      code: string
      sort: number
      /** 0 禁用 / 1 启用 */
      status: number
      /** 1 内置角色 / 2 自定义角色 */
      type: number
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 角色搜索参数 */
    interface RoleSearchParams {
      name?: string
      code?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑角色参数 */
    interface RoleSaveParams {
      name: string
      code: string
      sort?: number
      status: number
      remark?: string
    }

    /** 新增 / 编辑菜单参数 */
    interface MenuSaveParams {
      name: string
      permission?: string
      type: number
      sort?: number
      parentId?: string
      path?: string
      icon?: string
      component?: string
      componentName?: string
      status: number
      visible?: number
      keepAlive?: number
      alwaysShow?: number
    }

    /** 字典类型 */
    interface DictTypeItem {
      id: string
      name: string
      /** 字典类型编码，创建后不可修改 */
      type: string
      /** 0 禁用 / 1 启用 */
      status: number
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 字典类型列表 */
    interface DictTypeList {
      list: DictTypeItem[]
      total: number
    }

    /** 字典类型搜索参数 */
    interface DictTypeSearchParams {
      name?: string
      type?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑字典类型参数 */
    interface DictTypeSaveParams {
      name: string
      type: string
      status: number
      remark?: string
    }

    /** 字典数据 */
    interface DictDataItem {
      id: string
      sort: number
      label: string
      value: string
      dictType: string
      /** 0 禁用 / 1 启用 */
      status: number
      /** default / primary / success / info / warning / danger */
      colorType: string
      cssClass: string
      remark: string
      createdAt: string
      updatedAt: string
    }

    /** 字典数据列表 */
    interface DictDataList {
      list: DictDataItem[]
      total: number
    }

    /** 字典数据搜索参数 */
    interface DictDataSearchParams {
      dictType?: string
      label?: string
      status?: number
      pageNo?: number
      pageSize?: number
    }

    /** 新增 / 编辑字典数据参数 */
    interface DictDataSaveParams {
      sort?: number
      label: string
      value: string
      dictType: string
      status: number
      colorType?: string
      cssClass?: string
      remark?: string
    }

    /** 前端下拉 / 标签渲染用的字典项 */
    interface DictItem {
      label: string
      value: string
      colorType: string
      cssClass: string
      sort: number
    }

    /** 登录日志 */
    interface LoginLogItem {
      id: string
      userId: string
      /** 1 后台用户 / 2 会员 */
      userType: number
      username: string
      /** 1 登录 / 2 登出 */
      type: number
      /** 0 失败 / 1 成功 */
      result: number
      userIp: string
      userAgent: string
      createdAt: string
    }

    /** 登录日志列表 */
    interface LoginLogList {
      list: LoginLogItem[]
      total: number
    }

    /** 登录日志搜索参数 */
    interface LoginLogSearchParams {
      username?: string
      type?: number
      result?: number
      pageNo?: number
      pageSize?: number
    }

    /** 操作日志 */
    interface OperateLogItem {
      id: string
      userId: string
      /** 1 后台用户 / 2 会员 */
      userType: number
      /** 操作类型（模块） */
      type: string
      /** 操作子类型（动作） */
      subType: string
      action: string
      /** 0 失败 / 1 成功 */
      success: number
      extra: string
      requestMethod: string
      requestUrl: string
      requestIp: string
      userAgent: string
      /** 耗时（毫秒） */
      duration: number
      createdAt: string
    }

    /** 操作日志列表 */
    interface OperateLogList {
      list: OperateLogItem[]
      total: number
    }

    /** 操作日志搜索参数 */
    interface OperateLogSearchParams {
      type?: string
      subType?: string
      keyword?: string
      success?: number
      pageNo?: number
      pageSize?: number
    }

    /** 角色列表 */
    type RoleList = Api.Common.PaginatedResponse<RoleListItem>

    /** 角色列表项 */
    interface RoleListItem {
      roleId: number
      roleName: string
      roleCode: string
      description: string
      enabled: boolean
      createTime: string
    }

    /** 角色搜索参数 */
    type RoleSearchParams = Partial<
      Pick<RoleListItem, 'roleId' | 'roleName' | 'roleCode' | 'description' | 'enabled'> &
        Api.Common.CommonSearchParams
    >
  }
}
