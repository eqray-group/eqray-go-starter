<!-- 基础设施 / 定时任务：启停、改频率、手动执行与执行日志 -->
<template>
  <div class="infra-job-page art-full-height">
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
          <ElSpace wrap>
            <ElButton
              v-auth="'infra:job:update'"
              type="primary"
              @click="showDialog('add')"
              v-ripple
            >
              新增任务
            </ElButton>
            <ElTag type="info" effect="plain">调度器每分钟扫描一次，按 cron 命中执行</ElTag>
          </ElSpace>
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

    <JobFormDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :job-data="currentJob"
      :handlers="handlers"
      @success="refreshData"
    />

    <JobLogsDrawer v-model:visible="logsVisible" :job="currentJob" />
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchDeleteJob,
    fetchGetJobHandlers,
    fetchGetJobList,
    fetchRunJob,
    fetchUpdateJobStatus
  } from '@/api/infra/job'
  import JobFormDialog from './modules/job-form-dialog.vue'
  import JobLogsDrawer from './modules/job-logs-drawer.vue'

  defineOptions({ name: 'InfraJob' })

  type JobItem = Api.Infra.JobItem
  type HandlerItem = Api.Infra.JobHandlerItem
  type DialogType = 'add' | 'edit'

  const searchForm = ref<Api.Infra.JobSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '名称 / 标识',
      key: 'keyword',
      type: 'input',
      placeholder: '如 清理 / pay:sync-refunding',
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '启用', value: 1 },
          { label: '停用', value: 0 }
        ]
      }
    }
  ]

  const handlers = ref<HandlerItem[]>([])
  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentJob = ref<JobItem | null>(null)
  const logsVisible = ref(false)

  const formatDuration = (ms?: number): string => {
    if (!ms) return '0 ms'
    if (ms < 1000) return `${ms} ms`

    return `${(ms / 1000).toFixed(2)} s`
  }

  // 状态列：停用 / 启用
  const renderStatus = (row: JobItem) =>
    h(ElTag, { type: row.status === 1 ? 'success' : 'info', size: 'small' }, () =>
      row.status === 1 ? '启用' : '停用'
    )

  // 上次执行：时间 + 结果 + 耗时
  const renderLastRun = (row: JobItem) => {
    if (!row.lastRunAt) return h('span', { class: 'text-xs text-g-600' }, '尚未执行')

    return h('div', { class: 'flex flex-col gap-1' }, [
      h('span', { class: 'text-xs' }, row.lastRunAt),
      h('div', { class: 'flex items-center gap-2' }, [
        h(ElTag, { type: row.lastStatus === 1 ? 'success' : 'danger', size: 'small' }, () =>
          row.lastStatus === 1 ? '成功' : '失败'
        ),
        h('span', { class: 'text-xs text-g-600' }, formatDuration(row.lastDuration))
      ])
    ])
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
      apiFn: fetchGetJobList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'name',
          label: '任务',
          minWidth: 200,
          formatter: (row: JobItem) =>
            h('div', { class: 'flex flex-col' }, [
              h('span', { class: 'font-medium' }, row.name),
              h('span', { class: 'font-mono text-xs text-g-600' }, row.handler)
            ])
        },
        {
          prop: 'cron',
          label: '执行频率',
          width: 140,
          formatter: (row: JobItem) => h('span', { class: 'font-mono text-xs' }, row.cron)
        },
        { prop: 'status', label: '状态', width: 90, formatter: renderStatus },
        { prop: 'lastRunAt', label: '上次执行', width: 190, formatter: renderLastRun },
        {
          prop: 'nextRunAt',
          label: '下次执行',
          width: 165,
          formatter: (row: JobItem) =>
            h(
              'span',
              { class: row.nextRunAt ? 'text-xs' : 'text-xs text-g-600' },
              row.nextRunAt || (row.status === 1 ? '待计算' : '已停用')
            )
        },
        {
          prop: 'operation',
          label: '操作',
          width: 230,
          fixed: 'right',
          formatter: (row: JobItem) =>
            h('div', { class: 'flex items-center' }, [
              h(
                ElButton,
                {
                  link: true,
                  type: 'primary',
                  disabled: row.status !== 1,
                  onClick: () => handleRun(row)
                },
                () => '执行'
              ),
              h(
                ElButton,
                { link: true, type: 'primary', onClick: () => showLogs(row) },
                () => '日志'
              ),
              h(
                ElButton,
                {
                  link: true,
                  type: 'primary',
                  onClick: () => showDialog('edit', row)
                },
                () => '编辑'
              ),
              h(ElButton, { link: true, onClick: () => handleToggle(row) }, () =>
                row.status === 1 ? '停用' : '启用'
              ),
              h(
                ElButton,
                { link: true, type: 'danger', onClick: () => handleDelete(row) },
                () => '删除'
              )
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.JobSearchParams): void => {
    Object.assign(searchForm.value, params)
    Object.assign(searchParams, params)
    getData()
  }

  const loadHandlers = async (): Promise<void> => {
    try {
      handlers.value = (await fetchGetJobHandlers()) ?? []
    } catch {
      handlers.value = []
    }
  }

  const showDialog = (type: DialogType, row?: JobItem): void => {
    dialogType.value = type
    currentJob.value = row ?? null
    dialogVisible.value = true
  }

  const showLogs = (row: JobItem): void => {
    currentJob.value = row
    logsVisible.value = true
  }

  // 手动执行：同步等待结果，成功失败都用日志里的状态提示
  const handleRun = async (row: JobItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `立即执行「${row.name}」？手动执行不受 cron 影响，会在当前请求内跑完。`,
        '手动执行',
        { confirmButtonText: '执行', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      const log = await fetchRunJob(row.id)
      if (log?.status === 1) {
        ElMessage.success(`执行成功：${log.output || '完成'}`)
      } else {
        ElMessage.error(`执行失败：${log?.error || '未知原因'}`)
      }
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleToggle = async (row: JobItem): Promise<void> => {
    const next = row.status === 1 ? 0 : 1
    try {
      await fetchUpdateJobStatus(row.id, next)
      ElMessage.success(next === 1 ? '已启用' : '已停用')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleDelete = async (row: JobItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `删除任务「${row.name}」会同时清掉它的执行日志，确定删除？`,
        '删除任务',
        { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchDeleteJob(row.id)
      ElMessage.success('已删除')
      refreshData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(loadHandlers)
</script>
