<!-- 基础设施 / 省市县：查看已落库的行政区划，并在此触发全量同步 -->
<template>
  <div class="infra-region-page art-full-height">
    <ElCard class="mb-4 region-header-card">
      <div class="stat-row">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div class="flex flex-wrap items-center gap-8">
            <div v-for="item in statCards" :key="item.label" class="stat-item">
              <div class="stat-value">{{ item.value }}</div>
              <div class="stat-label">{{ item.label }}</div>
            </div>

            <ElDivider direction="vertical" class="!h-10" />

            <div class="text-xs text-g-600"> 最后同步：{{ stats.lastSyncedAt || '从未同步' }} </div>
          </div>

          <ElButton
            v-auth="'infra:region:sync'"
            type="primary"
            :loading="syncing"
            @click="handleSync"
            v-ripple
          >
            同步行政区划
          </ElButton>
        </div>
      </div>

      <ElAlert
        class="region-alert"
        type="info"
        :closable="false"
        show-icon
        title="数据源为高德行政区划，同步后整表刷新 infra_region"
      >
        <span class="text-xs">
          需要先在「参数配置」填写高德「Web 服务
          Key」（map.amap.serverKey）；前端省市区级联只查库、不再实时调高德。
        </span>
      </ElAlert>
    </ElCard>

    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElTag type="info" effect="plain"> 共 {{ stats.total }} 条，按区划编码升序 </ElTag>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElDivider,
    ElMessage,
    ElMessageBox,
    ElTag
  } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetInfraRegionList, fetchSyncInfraRegions } from '@/api/infra/region'
  import { clearRegionCache } from '@/utils/map/region'

  defineOptions({ name: 'InfraRegion' })

  type RegionRowItem = Api.Infra.RegionRowItem

  const LEVEL_TEXT: Record<string, string> = {
    province: '省',
    city: '市',
    district: '区县'
  }
  const LEVEL_TAG: Record<string, 'success' | 'warning' | 'info'> = {
    province: 'warning',
    city: 'success',
    district: 'info'
  }

  const searchForm = ref<Api.Infra.RegionSearchParams>({
    keyword: undefined,
    level: undefined,
    parentAdcode: undefined
  })

  const searchItems = [
    {
      label: '名称 / 编码',
      key: 'keyword',
      type: 'input',
      placeholder: '如 杭州 / 330100',
      clearable: true
    },
    {
      label: '层级',
      key: 'level',
      type: 'select',
      props: {
        placeholder: '全部层级',
        clearable: true,
        options: [
          { label: '省', value: 'province' },
          { label: '市', value: 'city' },
          { label: '区县', value: 'district' }
        ]
      }
    },
    {
      label: '父级编码',
      key: 'parentAdcode',
      type: 'input',
      placeholder: '如 330000',
      clearable: true
    }
  ]

  const stats = reactive<Api.Infra.RegionStats>({
    total: 0,
    province: 0,
    city: 0,
    district: 0,
    lastSyncedAt: ''
  })

  const statCards = computed(() => [
    { label: '总条数', value: stats.total },
    { label: '省', value: stats.province },
    { label: '市', value: stats.city },
    { label: '区县', value: stats.district }
  ])

  const syncing = ref(false)

  // 列表接口顺带返回统计，借 apiFn 顺手写入顶部卡片，不用多打一次接口
  const fetchRegionPage = async (params: Api.Infra.RegionSearchParams) => {
    const res = await fetchGetInfraRegionList(params)
    if (res?.stats) Object.assign(stats, res.stats)

    return res
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchRegionPage,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'adcode', label: '区划编码', width: 130 },
        { prop: 'name', label: '名称', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'level',
          label: '层级',
          width: 100,
          formatter: (row: RegionRowItem) =>
            h(
              ElTag,
              { type: LEVEL_TAG[row.level] || 'info', size: 'small' },
              () => LEVEL_TEXT[row.level] || row.level
            )
        },
        { prop: 'parentAdcode', label: '父级编码', width: 130 },
        { prop: 'sort', label: '排序', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: RegionRowItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info', size: 'small' }, () =>
              row.status === 1 ? '启用' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.RegionSearchParams): void => {
    Object.assign(searchForm.value, params)
    Object.assign(searchParams, params)
    getData()
  }

  // 全量同步：整表重刷，完成后清掉前端级联缓存并刷新列表
  const handleSync = async (): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        '将从高德重新拉取全国省、市、区县数据并覆盖 infra_region 表（约 3000+ 条）。' +
          '请确认已在「参数配置」填写高德 Web 服务 Key（map.amap.serverKey）。',
        '同步行政区划',
        { confirmButtonText: '开始同步', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    syncing.value = true
    try {
      const res = await fetchSyncInfraRegions()
      clearRegionCache()
      ElMessage.success(`同步完成，共 ${res?.count ?? 0} 条`)
      await getData()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      syncing.value = false
    }
  }
</script>

<style lang="scss" scoped>
  // 顶部卡片：数字区与提示条之间留出呼吸空间
  .region-header-card {
    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .stat-row {
    padding-bottom: 4px;
  }

  .stat-item {
    min-width: 76px;
  }

  .stat-value {
    font-size: 22px;
    font-weight: 600;
    line-height: 1.15;
    color: var(--el-text-color-primary);
  }

  .stat-label {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
  }

  .region-alert {
    margin-top: 20px;
    border-radius: 8px;

    :deep(.el-alert__content) {
      padding: 4px 0;
      line-height: 1.7;
    }
  }
</style>
