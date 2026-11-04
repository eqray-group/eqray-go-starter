/**
 * 通用类型（分页参数、统一响应结构）接口类型定义
 *
 * 全局命名空间 Api 的子命名空间，与其他模块的类型文件自动合并，业务代码无需导入即可使用：
 * `const item: Api.Common.XxxItem = ...`
 *
 * @module types/api/common
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

    /** 省市区级联选择结果；adcode 与 member_address.area_id 同源 */
    interface RegionValue {
      /** 省名称 */
      province: string
      /** 市名称（直辖市为空） */
      city: string
      /** 区 / 县名称 */
      district: string
      /** 省 adcode */
      provinceCode: string
      /** 市 adcode（直辖市为空） */
      cityCode: string
      /** 区县 adcode */
      districtCode: string
      /** 区县 adcode 的数字形式，可直接写入 member_address.area_id */
      areaId: number
    }

    /** 地图选点结果：经纬度为高德坐标系（GCJ-02），可直接用于高德地图展示 */
    interface MapLocation {
      /** 经度 */
      longitude: number
      /** 纬度 */
      latitude: number
      /** 格式化地址（逆地理编码结果） */
      address: string
      /** 省 */
      province?: string
      /** 市 */
      city?: string
      /** 区 / 县 */
      district?: string
      /** 区县 adcode，可直接写入 member_address.area_id */
      areaId?: number
    }
  }
}
