<!--
  省市区级联选择组件

  数据来自高德 DistrictSearch（与地图选点共用后台配置的 Key），逐级懒加载 + 本地缓存 7 天，
  无需引入 pca.json / china-division 之类的数据包，行政区划调整也不用发版。
  选完返回省 / 市 / 区的名称与 adcode，其中 areaId 可直接写入 member_address.area_id。

  用法：
    <RegionCascader v-model="region" placeholder="请选择省 / 市 / 区" />
    const region = ref<Api.Common.RegionValue | null>(null)
    // { province: '浙江省', city: '杭州市', district: '西湖区',
    //   provinceCode: '330000', cityCode: '330100', districtCode: '330106', areaId: 330106 }

  只要省市级联（不选到区）时传 :show-district="false"。
-->
<template>
  <ElCascader
    ref="cascaderRef"
    :model-value="codes"
    :props="cascaderProps"
    :placeholder="placeholder"
    :disabled="disabled"
    :clearable="clearable"
    :show-all-levels="true"
    class="w-full"
    @change="handleChange"
  />
</template>

<script setup lang="ts">
  import type { CascaderProps } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import {
    CHINA_ADCODE,
    clearRegionCache,
    loadRegionChildren,
    type RegionLevel,
    type RegionOption
  } from '@/utils/map/region'

  defineOptions({ name: 'RegionCascader' })

  type RegionValue = Api.Common.RegionValue

  /** ElCascader 实例只用到 getCheckedNodes，避免引入内部类型 */
  interface CascaderNodeLike {
    value?: string
    label?: string
    pathLabels?: string[]
  }

  interface Props {
    /** 选中的省市区，可用 v-model 双向绑定 */
    modelValue?: RegionValue | null
    /** 未选择时的提示文案 */
    placeholder?: string
    /** 是否禁用 */
    disabled?: boolean
    /** 是否可清空 */
    clearable?: boolean
    /** 是否选到区县；false 时只到市 */
    showDistrict?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: null,
    placeholder: '请选择省 / 市 / 区',
    disabled: false,
    clearable: true,
    showDistrict: true
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: RegionValue | null): void
    (e: 'change', value: RegionValue | null): void
  }>()

  const cascaderRef = ref<{ getCheckedNodes: (leafOnly?: boolean) => CascaderNodeLike[] }>()

  /** 已加载过的区划登记表：code → { label, level }，用于把 codes 还原成结构化结果 */
  const registry = new Map<string, RegionOption>()

  /** 已提示过加载失败，避免展开每一级都弹一次 */
  let warned = false

  // 级联显示值：adcode 数组，父组件只传结构化对象，这里按层级还原
  const codes = computed(() => {
    const value = props.modelValue
    if (!value) {
      return []
    }

    return [value.provinceCode, value.cityCode, value.districtCode].filter((code): code is string =>
      Boolean(code)
    )
  })

  // 用框架的类型约束 props，节点结构由 element-plus 提供，避免自己声明导致不兼容
  const cascaderProps: CascaderProps = {
    lazy: true,
    lazyLoad: (node, resolve) => {
      void loadChildren(node, resolve)
    },
    value: 'value',
    label: 'label',
    leaf: 'leaf',
    expandTrigger: 'click'
  }

  /**
   * 懒加载下一级
   * 根节点（level = 0）从全国开始，其余用当前节点的 adcode 与层级去查子级
   */
  async function loadChildren(
    node: { level?: number; value?: unknown; data?: unknown },
    resolve: (nodes: RegionOption[]) => void
  ): Promise<void> {
    const isRoot = !node?.level
    const parentCode = isRoot ? CHINA_ADCODE : String(node.value ?? CHINA_ADCODE)
    const parentLevel: RegionLevel = isRoot
      ? 'country'
      : ((node.data as RegionOption | undefined)?.level ?? 'country')

    // 只要到市：市级节点直接标记为叶子，不再请求下级
    if (!props.showDistrict && parentLevel === 'city') {
      resolve([])

      return
    }

    try {
      const children = await loadRegionChildren(parentCode)
      children.forEach((item) => registry.set(item.value, item))

      resolve(
        props.showDistrict
          ? children
          : children.map((item) => ({
              ...item,
              leaf: item.leaf || item.level === 'city'
            }))
      )
    } catch (error) {
      if (!warned) {
        warned = true
        ElMessage.warning((error as Error).message)
      }
      resolve([])
    }
  }

  /**
   * 选中变化：按 adcode 对应的层级归类，避免直辖市少一级导致省市错位
   */
  const handleChange = (value: unknown): void => {
    const selected = Array.isArray(value) ? (value as string[]) : []
    if (!selected.length) {
      emit('update:modelValue', null)
      emit('change', null)

      return
    }

    const pathLabels = cascaderRef.value?.getCheckedNodes(true)?.[0]?.pathLabels ?? []

    let province = ''
    let city = ''
    let district = ''
    let provinceCode = ''
    let cityCode = ''
    let districtCode = ''

    selected.forEach((code, index) => {
      const meta = registry.get(code)
      // 兜底：拿不到 level 时按 省 → 市 → 区 的层级顺序推断
      const level: RegionLevel =
        meta?.level ?? (index === 0 ? 'province' : index === 1 ? 'city' : 'district')
      const label = meta?.label ?? pathLabels[index] ?? ''

      if (level === 'province' || level === 'country') {
        province = label
        provinceCode = code
        return
      }
      if (level === 'city') {
        city = label
        cityCode = code
        return
      }

      district = label
      districtCode = code
    })

    const result: RegionValue = {
      province,
      city,
      district,
      provinceCode,
      cityCode,
      districtCode,
      areaId: Number(districtCode || cityCode || provinceCode || 0)
    }

    emit('update:modelValue', result)
    emit('change', result)
  }

  defineExpose({
    /** 强制清空行政区划缓存后重新拉取 */
    refresh: clearRegionCache
  })
</script>
