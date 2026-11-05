<!-- 定时任务执行日志：按任务维度查看，展开行看输出与错误 -->
<template>
  <ElDrawer v-model="drawerVisible" :title="`执行日志 · ${job?.name || ''}`" size="720px">
    <div class="mb-3 flex items-center gap-2">
      <ElSelect
        v-model="status"
        placeholder="全部结果"
        clearable
        style="width: 140px"
        @change="handleSearch"
      >
        <ElOption label="成功" :value="1" />
        <ElOption label="失败" :value="0" />
      </ElSelect>
      <ElButton @click="loadLogs">刷新</ElButton>
      <span class="text-xs text-g-600">仅保留最近 30 天</span>
    </div>

    <ElTable v-loading="loading" :data="list" border>
      <ElTableColumn type="expand">
        <template #default="{ row }">
          <div class="log-detail">
            <div v-if="row.error" class="mb-2">
              <div class="log-label">失败原因</div>
              <pre class="log-pre is-error">{{ row.error }}</pre>
            </div>
            <div>
              <div class="log-label">执行输出</div>
              <pre class="log-pre">{{ row.output || '（无输出）' }}</pre>
            </div>
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="startedAt" label="开始时间" width="165" />
      <ElTableColumn label="触发" width="80">
        <template #default="{ row }">
          <ElTag :type="row.trigger === 'manual' ? 'warning' : 'info'" size="small">
            {{ row.trigger === 'manual' ? '手动' : '调度' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="结果" width="90">
        <template #default="{ row }">
          <ElTag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.status === 1 ? '成功' : '失败' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="耗时" width="100">
        <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
      </ElTableColumn>
      <ElTableColumn prop="output" label="输出摘要" min-width="180" show-overflow-tooltip />
      <template #empty>
        <ElEmpty description="暂无执行记录" />
      </template>
    </ElTable>

    <div class="mt-4 flex justify-end">
      <ElPagination
        v-model:current-page="pageNo"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadLogs"
        @size-change="loadLogs"
      />
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import {
    ElButton,
    ElDrawer,
    ElEmpty,
    ElOption,
    ElPagination,
    ElSelect,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import { fetchGetJobLogs } from '@/api/infra/job'

  defineOptions({ name: 'JobLogsDrawer' })

  type JobItem = Api.Infra.JobItem
  type JobLogItem = Api.Infra.JobLogItem

  interface Props {
    visible: boolean
    job?: JobItem | null
  }

  const props = withDefaults(defineProps<Props>(), { job: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const list = ref<JobLogItem[]>([])
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = ref(20)
  const status = ref<number | undefined>(undefined)

  const formatDuration = (ms?: number): string => {
    if (!ms) return '0 ms'
    if (ms < 1000) return `${ms} ms`

    return `${(ms / 1000).toFixed(2)} s`
  }

  const loadLogs = async (): Promise<void> => {
    if (!props.job) return

    loading.value = true
    try {
      const res = await fetchGetJobLogs({
        jobId: props.job.id,
        status: status.value,
        pageNo: pageNo.value,
        pageSize: pageSize.value
      })
      list.value = res?.list ?? []
      total.value = res?.total ?? 0
    } catch {
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (): void => {
    pageNo.value = 1
    loadLogs()
  }

  watch(drawerVisible, (visible) => {
    if (!visible) return
    pageNo.value = 1
    loadLogs()
  })
</script>

<style lang="scss" scoped>
  .log-detail {
    padding: 8px 16px;
  }

  .log-label {
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .log-pre {
    max-height: 200px;
    padding: 10px 12px;
    margin: 0;
    overflow: auto;
    font-size: 12px;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--el-fill-color-lighter);
    border-radius: 6px;

    &.is-error {
      color: var(--el-color-danger);
    }
  }
</style>
