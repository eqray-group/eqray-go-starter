<!--
  高德地图选点组件

  点输入框打开地图，支持「关键字搜索」与「地图点选」，确认后回填经纬度 + 地址（含省市区）。
  Key 不写死在前端：运行时从后台「基础设施 / 参数配置」读取 map.amap.key 与 map.amap.securityCode，
  换 Key、切环境都不需要重新打包。

  用法：
    <AmapPicker v-model="location" placeholder="请选择收货地址" />
    const location = ref<Api.Common.MapLocation | null>(null)
    // location 形如 { longitude: 121.47, latitude: 31.23, address: '上海市黄浦区…', province, city, district }
-->
<template>
  <div class="amap-picker">
    <ElInput
      :model-value="displayValue"
      :placeholder="placeholder"
      :disabled="disabled"
      readonly
      clearable
      class="cursor-pointer"
      @click="open"
      @clear="handleClear"
    >
      <template #suffix>
        <ArtSvgIcon icon="ri:map-pin-line" class="text-g-500" />
      </template>
    </ElInput>

    <ElDialog
      v-model="visible"
      title="选择位置"
      width="880px"
      align-center
      :close-on-click-modal="false"
      @opened="initMap"
      @closed="handleClosed"
    >
      <div class="mb-3 flex items-center gap-2">
        <ElInput
          v-model="keyword"
          placeholder="搜索地点，如：人民广场"
          clearable
          @keyup.enter="handleSearch"
        />
        <ElButton type="primary" :loading="searching" @click="handleSearch">搜索</ElButton>
      </div>

      <div v-if="tips.length" class="tip-list">
        <div
          v-for="(tip, index) in tips"
          :key="`${tip.name}-${index}`"
          class="tip-item"
          @click="handlePickTip(tip)"
        >
          <span class="truncate">{{ tip.name }}</span>
          <span class="ml-2 shrink-0 text-xs text-g-600">{{ tip.district }}</span>
        </div>
      </div>

      <div ref="mapRef" class="amap-container" :style="{ height }" />

      <div class="mt-3 flex items-center justify-between text-sm">
        <span v-if="picked?.address" class="text-g-600">
          已选：{{ picked.address }}（{{ picked.longitude.toFixed(6) }},
          {{ picked.latitude.toFixed(6) }}）
        </span>
        <span v-else class="text-g-600">在地图上点击，或搜索地点后选择</span>
        <span v-if="errorMsg" class="text-danger">{{ errorMsg }}</span>
      </div>

      <template #footer>
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" :disabled="!picked?.address" @click="handleConfirm">确定</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { ElDialog, ElMessage } from 'element-plus'
  import { ensureAmapPlugins, loadAmap } from '@/utils/map/amap'

  defineOptions({ name: 'AmapPicker' })

  type MapLocation = Api.Common.MapLocation

  interface Props {
    /** 选中的位置，可用 v-model 双向绑定 */
    modelValue?: MapLocation | null
    /** 未选择时的提示文案 */
    placeholder?: string
    /** 是否禁用 */
    disabled?: boolean
    /** 地图容器高度 */
    height?: string
    /** 搜索建议的偏向城市，默认全国 */
    city?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: null,
    placeholder: '请在地图上选择位置',
    disabled: false,
    height: '420px',
    city: ''
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: MapLocation | null): void
    (e: 'change', value: MapLocation | null): void
  }>()

  /** 默认中心点：北京天安门，兜底用 */
  const DEFAULT_CENTER: [number, number] = [116.397428, 39.90923]

  const visible = ref(false)
  const keyword = ref('')
  const searching = ref(false)
  const errorMsg = ref('')
  const tips = ref<AMap.Tip[]>([])
  const picked = ref<MapLocation | null>(null)
  const mapRef = ref<HTMLElement>()

  let mapInstance: AMap.Map | null = null
  let markerInstance: AMap.Marker | null = null
  let geocoder: AMap.Geocoder | null = null
  let autoComplete: AMap.AutoComplete | null = null

  // 输入框展示地址；坐标对用户没意义，只在弹窗里展示
  const displayValue = computed(() => props.modelValue?.address ?? '')

  // 打开地图：先把当前值带进去，避免每次都要重新找
  const open = (): void => {
    if (props.disabled) {
      return
    }

    picked.value = props.modelValue ? { ...props.modelValue } : null
    visible.value = true
  }

  // 初始化地图与插件：只在弹窗打开后执行，避免页面加载就拉起第三方脚本
  const initMap = async (): Promise<void> => {
    errorMsg.value = ''

    try {
      await loadAmap()
      await ensureAmapPlugins()
    } catch (error) {
      errorMsg.value = (error as Error).message

      return
    }

    const AMapNS = window.AMap
    if (!AMapNS || !mapRef.value) {
      return
    }

    const center: [number, number] = picked.value
      ? [picked.value.longitude, picked.value.latitude]
      : DEFAULT_CENTER
    const zoom = picked.value ? 15 : 12

    if (mapInstance) {
      mapInstance.setCenter(center)
      mapInstance.setZoom(zoom)
      markerInstance?.setPosition(center)

      return
    }

    mapInstance = new AMapNS.Map(mapRef.value, { zoom, center, viewMode: '2D' })
    markerInstance = new AMapNS.Marker({ position: center })
    mapInstance.add(markerInstance)
    mapInstance.on('click', (event) => {
      void pickByLngLat(event.lnglat.getLng(), event.lnglat.getLat())
    })

    geocoder = new AMapNS.Geocoder()
    autoComplete = new AMapNS.AutoComplete({ city: props.city || '全国' })
  }

  // 落点 + 逆地理编码拿结构化地址
  const pickByLngLat = async (longitude: number, latitude: number): Promise<void> => {
    picked.value = { longitude, latitude, address: '' }
    markerInstance?.setPosition([longitude, latitude])
    mapInstance?.setCenter([longitude, latitude])

    if (!geocoder) {
      return
    }

    await new Promise<void>((resolve) => {
      geocoder?.getAddress([longitude, latitude], (status, result) => {
        if (status === 'complete' && result.regeocode) {
          const component = result.regeocode.addressComponent ?? {}
          picked.value = {
            longitude,
            latitude,
            address: result.regeocode.formattedAddress ?? `${longitude},${latitude}`,
            province: component.province,
            city: component.city,
            district: component.district,
            areaId: Number(component.adcode || 0)
          }
        }
        resolve()
      })
    })
  }

  // 关键字搜索：拿到建议列表供用户点选
  const handleSearch = (): void => {
    const value = keyword.value.trim()
    if (!value) {
      return
    }

    if (!autoComplete) {
      ElMessage.warning('地图尚未加载完成，请稍后再试')

      return
    }

    searching.value = true
    autoComplete.search(value, (status, result) => {
      searching.value = false
      tips.value = status === 'complete' ? (result.tips ?? []).filter((tip) => tip.location) : []
      if (!tips.value.length) {
        ElMessage.warning('没有找到匹配的地点')
      }
    })
  }

  // 选中建议项：需要 location（没有坐标的条目无法落点）
  const handlePickTip = async (tip: AMap.Tip): Promise<void> => {
    tips.value = []
    if (!tip.location) {
      return
    }

    await pickByLngLat(tip.location.lng, tip.location.lat)
  }

  // 确认回填：地址取不到时不允许提交，避免只存了坐标
  const handleConfirm = (): void => {
    if (!picked.value?.address) {
      return
    }

    const value: MapLocation = { ...picked.value }
    emit('update:modelValue', value)
    emit('change', value)
    visible.value = false
  }

  // 清空：交给父组件决定是否允许为空
  const handleClear = (): void => {
    emit('update:modelValue', null)
    emit('change', null)
  }

  // 关闭后销毁地图实例，避免反复打开积累图表与事件监听
  const handleClosed = (): void => {
    mapInstance?.destroy()
    mapInstance = null
    markerInstance = null
    geocoder = null
    autoComplete = null
    tips.value = []
    keyword.value = ''
  }

  defineExpose({ open })
</script>

<style lang="scss" scoped>
  .amap-container {
    width: 100%;
    border-radius: 8px;
    overflow: hidden;
    background: var(--el-fill-color-light);
  }

  .tip-list {
    max-height: 180px;
    margin-bottom: 12px;
    overflow: auto;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
  }

  .tip-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    font-size: 13px;
    cursor: pointer;
    transition: background-color 0.2s;

    &:hover {
      background: var(--el-fill-color-light);
    }
  }
</style>
