/**
 * 登录认证与会话接口类型定义
 *
 * 全局命名空间 Api 的子命名空间，与其他模块的类型文件自动合并，业务代码无需导入即可使用：
 * `const item: Api.Auth.XxxItem = ...`
 *
 * @module types/api/auth
 */

declare namespace Api {
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
}
