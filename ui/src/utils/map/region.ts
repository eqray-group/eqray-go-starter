/**
 * 省市区数据源（数据库版）
 *
 * 数据由后台一次性同步到 `infra_region` 表（数据源是高德行政区划接口，由后端拉取），
 * 前端只调 `GET /api/v1/client/infra/regions` 按父级取下一级，**不再直连高德**：
 * - 页面秒开，不受高德接口配额 / 网络抖动影响；
 * - 行政区划变动时在后台重新同步即可，前端无需发版；
 * - 两级缓存（内存 + localStorage 30 天），同一父级只会请求一次。
 *
 * 如果后台还没同步过数据，接口会返回「行政区划数据未初始化」的提示。
 *
 * @module utils/map/region
 */

import { fetchGetRegionChildren } from '@/api/infra/region'

/** 行政区划层级 */
export type RegionLevel = 'country' | 'province' | 'city' | 'district' | 'street'

/** 级联选择的一项，字段名与 ElCascader 的 props 对应 */
export interface RegionOption {
  /** 区划编码 adcode */
  value: string
  /** 区划名称 */
  label: string
  /** 层级 */
  level: RegionLevel
  /** 是否叶子节点（区县为 true） */
  leaf: boolean
  /** 允许携带 element-plus 级联节点的扩展字段，保持与 CascaderOption 兼容 */
  [key: string]: unknown
}

/** 全国行政区划编码，作为级联根节点 */
export const CHINA_ADCODE = '100000'

/** 缓存有效期：30 天。数据在库里且很少变动，过期后重新拉一次即可 */
const CACHE_TTL = 30 * 24 * 60 * 60 * 1000
const CACHE_PREFIX = 'app:region:v2:'

/** 进程内缓存：同一页面多次展开同一级不再请求 */
const memoryCache = new Map<string, RegionOption[]>()

/**
 * 加载某个区划的下级列表
 * @param parentCode 父级 adcode；根节点传 CHINA_ADCODE
 * @returns 下级区划列表，无下级时返回空数组
 */
export async function loadRegionChildren(
  parentCode: string = CHINA_ADCODE
): Promise<RegionOption[]> {
  const key = parentCode || CHINA_ADCODE

  const fromMemory = memoryCache.get(key)
  if (fromMemory) {
    return fromMemory
  }

  const fromStorage = readCache(key)
  if (fromStorage) {
    memoryCache.set(key, fromStorage)

    return fromStorage
  }

  const list = await fetchGetRegionChildren(key)
  const options: RegionOption[] = (list ?? []).map((item) => ({
    value: item.adcode,
    label: item.name,
    level: item.level as RegionLevel,
    leaf: item.leaf
  }))

  memoryCache.set(key, options)
  writeCache(key, options)

  return options
}

/**
 * 清空省市区缓存：后台重新同步行政区划后调用，让下次展开拉取最新数据
 */
export function clearRegionCache(): void {
  memoryCache.clear()

  try {
    Object.keys(localStorage)
      .filter((key) => key.startsWith(CACHE_PREFIX))
      .forEach((key) => localStorage.removeItem(key))
  } catch {
    // 隐私模式下 localStorage 不可用，忽略
  }
}

/**
 * 读缓存；过期或结构不合法视为未命中
 */
function readCache(key: string): RegionOption[] | null {
  try {
    const raw = localStorage.getItem(CACHE_PREFIX + key)
    if (!raw) {
      return null
    }

    const parsed = JSON.parse(raw) as { at?: number; list?: RegionOption[] }
    if (!parsed?.at || !Array.isArray(parsed.list) || Date.now() - parsed.at > CACHE_TTL) {
      localStorage.removeItem(CACHE_PREFIX + key)

      return null
    }

    return parsed.list
  } catch {
    return null
  }
}

/**
 * 写缓存；写入失败（超限 / 隐私模式）不影响功能
 */
function writeCache(key: string, list: RegionOption[]): void {
  try {
    localStorage.setItem(CACHE_PREFIX + key, JSON.stringify({ at: Date.now(), list }))
  } catch {
    // 忽略：下次重新拉取即可
  }
}
