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
      email: string
      avatar?: string
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
