/**
 * 高德地图 JS API 加载器
 *
 * Key 不写死在前端：运行时从后台「基础设施 / 参数配置」读取
 * （`map.amap.key` 与 `map.amap.securityCode`），首次读取走免登录接口
 * `GET /api/v1/client/infra/public-configs`，结果缓存在 sessionStorage。
 *
 * 这样换 Key、切正式环境都只需要在后台改参数，不用重新打包前端。
 *
 * @module utils/map/amap
 */

import { fetchGetPublicConfigs } from '@/api/infra/public-config'

/** 高德地图运行时配置 */
export interface AmapRuntimeConfig {
  /** Web 端（JS API）Key */
  key: string
  /** 安全密钥 securityJsCode，老 Key 可为空 */
  securityCode: string
}

/** 参与读取的后台参数键，必须与后端 publicConfigKeys 白名单一致 */
const PUBLIC_CONFIG_KEYS = ['map.amap.key', 'map.amap.securityCode']
const STORAGE_KEY = 'app:amap:config'
const SCRIPT_ID = 'amap-jsapi'
const PLUGINS = ['AMap.Geocoder', 'AMap.AutoComplete']

let runtimeConfig: AmapRuntimeConfig | null = null
let configTask: Promise<AmapRuntimeConfig> | null = null
let loadedKey = ''
let pluginTask: Promise<void> | null = null

/**
 * 加载高德地图 JS API（幂等，可并发调用）
 * @param force 是否强制重新读取后台配置（后台改 Key 后可传 true）
 * @throws 后台未配置 Key 或脚本加载失败时抛出可展示的错误
 */
export async function loadAmap(force = false): Promise<void> {
  const config = await resolveConfig(force)

  if (window.AMap && loadedKey === config.key) {
    return
  }

  await injectScript(config)
  loadedKey = config.key
}

/**
 * 加载 Geocoder / AutoComplete 插件（2.0 默认只含核心库）
 */
export async function ensureAmapPlugins(): Promise<void> {
  if (pluginTask) {
    return pluginTask
  }

  pluginTask = new Promise<void>((resolve, reject) => {
    if (!window.AMap) {
      reject(new Error('高德地图尚未加载'))

      return
    }

    window.AMap.plugin(PLUGINS, () => resolve())
  })

  return pluginTask
}

/**
 * 清掉内存与 sessionStorage 里的 Key 缓存，下次加载重新读后台
 */
export function clearAmapCache(): void {
  runtimeConfig = null
  configTask = null
  loadedKey = ''
  pluginTask = null
  try {
    sessionStorage.removeItem(STORAGE_KEY)
  } catch {
    // 隐私模式下 sessionStorage 不可用，忽略即可
  }
}

/**
 * 解析运行时配置：内存 → sessionStorage → 后台参数
 */
async function resolveConfig(force: boolean): Promise<AmapRuntimeConfig> {
  if (!force) {
    if (runtimeConfig?.key) {
      return runtimeConfig
    }

    const stored = readStorage()
    if (stored?.key) {
      runtimeConfig = stored

      return stored
    }
  }

  if (!configTask) {
    configTask = fetchGetPublicConfigs(PUBLIC_CONFIG_KEYS)
      .then((values) => {
        const key = String(values?.['map.amap.key'] ?? '').trim()
        const securityCode = String(values?.['map.amap.securityCode'] ?? '').trim()

        if (!key) {
          throw new Error('后台未配置高德地图 Key，请到「基础设施 → 参数配置」填写 map.amap.key')
        }

        const config: AmapRuntimeConfig = { key, securityCode }
        runtimeConfig = config
        writeStorage(config)

        return config
      })
      .finally(() => {
        configTask = null
      })
  }

  return configTask
}

/**
 * 注入 JS API 脚本；Key 变化时先移除旧脚本再重新注入
 */
function injectScript(config: AmapRuntimeConfig): Promise<void> {
  if (config.securityCode) {
    window._AMapSecurityConfig = { securityJsCode: config.securityCode }
  }

  return new Promise<void>((resolve, reject) => {
    document.getElementById(SCRIPT_ID)?.remove()

    const script = document.createElement('script')
    script.id = SCRIPT_ID
    script.async = true
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${encodeURIComponent(config.key)}`
    script.onload = () => resolve()
    script.onerror = () =>
      reject(new Error('高德地图脚本加载失败，请检查 Key 是否正确、域名是否加入 Key 的白名单'))

    document.head.appendChild(script)
  })
}

/**
 * 读取 sessionStorage 中的配置
 */
function readStorage(): AmapRuntimeConfig | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return null
    }

    const parsed = JSON.parse(raw) as AmapRuntimeConfig

    return parsed?.key ? parsed : null
  } catch {
    return null
  }
}

/**
 * 写入 sessionStorage，避免每次渲染都请求后台
 */
function writeStorage(config: AmapRuntimeConfig): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(config))
  } catch {
    // 隐私模式下写入失败不影响功能，内存里已有配置
  }
}
