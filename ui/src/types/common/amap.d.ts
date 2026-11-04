/**
 * 高德地图 JS API 2.0 的最小类型声明
 *
 * 只声明地图选点组件实际用到的成员，避免引入 @types/amap-js-api 与
 * @amap/amap-jsapi-loader 这两个额外依赖（JS 脚本在运行时按后台配置的 Key 动态注入）。
 *
 * @module types/common/amap
 */

declare namespace AMap {
  /** 经纬度对象 */
  interface LngLat {
    lng: number
    lat: number
    getLng(): number
    getLat(): number
  }

  /** 地图实例 */
  interface Map {
    on(event: 'click', handler: (event: { lnglat: LngLat }) => void): void
    setCenter(center: [number, number]): void
    setZoom(zoom: number): void
    add(overlay: unknown): void
    remove(overlay: unknown): void
    setFitView(overlays?: unknown[]): void
    destroy(): void
  }

  /** 点标记 */
  interface Marker {
    setPosition(position: [number, number]): void
    setMap(map: Map | null): void
  }

  /** 逆地理编码返回的行政区划 */
  interface AddressComponent {
    province?: string
    city?: string
    district?: string
    citycode?: string
    /** 区县 adcode，可直接写入会员地址的 area_id */
    adcode?: string
  }

  /** 逆地理编码结果 */
  interface GeocodeResult {
    regeocode?: {
      formattedAddress?: string
      addressComponent?: AddressComponent
    }
  }

  /** 逆地理编码器 */
  interface Geocoder {
    getAddress(
      lnglat: [number, number],
      callback: (status: string, result: GeocodeResult) => void
    ): void
  }

  /** 搜索建议条目 */
  interface Tip {
    name: string
    district?: string
    adcode?: string
    location?: { lng: number; lat: number }
  }

  /** 搜索建议结果 */
  interface AutoCompleteResult {
    tips: Tip[]
  }

  /** 输入提示（搜索建议） */
  interface AutoComplete {
    search(keyword: string, callback: (status: string, result: AutoCompleteResult) => void): void
  }

  const Map: new (container: HTMLElement, options?: Record<string, unknown>) => Map
  const Marker: new (options?: Record<string, unknown>) => Marker
  const Geocoder: new (options?: Record<string, unknown>) => Geocoder
  const AutoComplete: new (options?: Record<string, unknown>) => AutoComplete
  const LngLat: new (lng: number, lat: number) => LngLat

  /** 加载插件（2.0 的 Geocoder / AutoComplete 需要显式加载） */
  function plugin(names: string[], callback?: () => void): void
}

declare interface Window {
  /** 高德 JS API 注入后挂载的命名空间 */
  AMap?: typeof AMap
  /** 2021-12 之后申请的 Key 需要配套的安全密钥配置 */
  _AMapSecurityConfig?: { securityJsCode?: string }
}
