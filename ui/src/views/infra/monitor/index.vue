<!-- 基础设施 / 系统监控：只读聚合指标，30s 自动刷新 -->
<template>
  <div class="infra-monitor-page art-full-height">
    <div class="monitor-toolbar">
      <div class="flex items-center gap-3">
        <ElTag v-if="data" :type="overallHealthy ? 'success' : 'danger'" effect="dark" size="large">
          <span class="status-dot" :class="overallHealthy ? 'is-ok' : 'is-bad'"></span>
          {{ overallHealthy ? '运行正常' : '存在异常依赖' }}
        </ElTag>
        <span v-if="data" class="text-xs text-g-600">采集于 {{ data.collectedAt }}</span>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-xs text-g-600">自动刷新</span>
        <ElSwitch v-model="autoRefresh" size="small" />
        <ElButton :loading="loading" @click="loadOverview">刷新</ElButton>
      </div>
    </div>

    <ElCard v-loading="loading" class="monitor-card">
      <div class="monitor-grid">
        <!-- 服务状态 -->
        <section class="stat-panel">
          <div class="panel-title">服务状态</div>
          <div class="panel-value">{{ formatUptime(data?.app.uptimeSeconds) }}</div>
          <div class="panel-sub">已运行时长</div>
          <div class="kv-list">
            <div class="kv"
              ><span>Go 版本</span><b>{{ data?.app.goVersion || '-' }}</b></div
            >
            <div class="kv"
              ><span>协程数</span><b>{{ data?.app.goroutines ?? '-' }}</b></div
            >
            <div class="kv"
              ><span>CPU 核数</span><b>{{ data?.app.cpuCores ?? '-' }}</b></div
            >
          </div>
        </section>

        <!-- 内存 -->
        <section class="stat-panel">
          <div class="panel-title">内存占用</div>
          <div class="panel-value">{{ data?.memory.allocMB ?? 0 }} MB</div>
          <div class="panel-sub">堆上活跃对象</div>
          <div class="progress-row">
            <div class="progress-track">
              <div class="progress-fill" :style="{ width: `${heapPercent}%` }"></div>
            </div>
            <span class="text-xs text-g-600">{{ heapPercent }}%</span>
          </div>
          <div class="kv-list">
            <div class="kv"
              ><span>堆使用</span><b>{{ data?.memory.heapInuseMB ?? 0 }} MB</b></div
            >
            <div class="kv"
              ><span>进程申请</span><b>{{ data?.memory.sysMB ?? 0 }} MB</b></div
            >
            <div class="kv"
              ><span>GC / 对象数</span
              ><b>{{ data?.memory.numGC ?? 0 }} / {{ data?.memory.heapObjects ?? 0 }}</b></div
            >
          </div>
        </section>

        <!-- 数据库 -->
        <section class="stat-panel" :class="healthClass(data?.db.ok)">
          <div class="panel-title">数据库</div>
          <div class="panel-value">{{ data?.db.ok ? '可用' : '异常' }}</div>
          <div class="panel-sub">探测延迟 {{ data?.db.latencyMs ?? 0 }} ms</div>
          <div v-if="data && !data.db.ok" class="panel-error">{{ data.db.message }}</div>
        </section>

        <!-- 缓存 -->
        <section class="stat-panel" :class="healthClass(data?.cache.ok)">
          <div class="panel-title">缓存</div>
          <div class="panel-value">{{ data?.cache.ok ? '可用' : '异常' }}</div>
          <div class="panel-sub">探测延迟 {{ data?.cache.latencyMs ?? 0 }} ms</div>
          <div v-if="data && !data.cache.ok" class="panel-error">{{ data.cache.message }}</div>
        </section>

        <!-- 近 24h 请求 -->
        <section class="stat-panel span-2">
          <div class="panel-title">近 24 小时请求</div>
          <div class="metric-row">
            <div class="metric">
              <div class="metric-value">{{ data?.request.total ?? 0 }}</div>
              <div class="metric-label">请求总数</div>
            </div>
            <div class="metric">
              <div class="metric-value" :class="{ 'is-danger': (data?.request.errors ?? 0) > 0 }">
                {{ data?.request.errors ?? 0 }}
              </div>
              <div class="metric-label">失败请求</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.request.slow ?? 0 }}</div>
              <div class="metric-label">慢请求（≥1s）</div>
            </div>
            <div class="metric">
              <div class="metric-value">{{ data?.request.avgDurationMs ?? 0 }} ms</div>
              <div class="metric-label">平均耗时</div>
            </div>
          </div>
        </section>

        <!-- 最近错误 -->
        <section class="stat-panel span-2 is-table">
          <div class="panel-title">
            最近错误
            <span class="ml-2 text-xs font-normal text-g-600">仅展示最近 10 条</span>
          </div>
          <ElTable :data="data?.recentErrors ?? []" size="small" max-height="320">
            <ElTableColumn prop="createdAt" label="时间" width="170" />
            <ElTableColumn prop="name" label="错误" width="160" show-overflow-tooltip />
            <ElTableColumn prop="requestUrl" label="接口" min-width="200" show-overflow-tooltip />
            <ElTableColumn prop="message" label="信息" min-width="260" show-overflow-tooltip />
            <template #empty>
              <ElEmpty description="近 24 小时没有错误日志" :image-size="80" />
            </template>
          </ElTable>
        </section>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElButton, ElCard, ElEmpty, ElSwitch, ElTable, ElTableColumn, ElTag } from 'element-plus'
  import { fetchGetMonitorOverview } from '@/api/infra/monitor'

  defineOptions({ name: 'InfraMonitor' })

  const loading = ref(false)
  const data = ref<Api.Infra.MonitorOverview | null>(null)
  const autoRefresh = ref(true)

  let timer: ReturnType<typeof setInterval> | null = null

  // 两个依赖都可算健康：DB + 缓存
  const overallHealthy = computed(
    () => data.value?.db.ok !== false && data.value?.cache.ok !== false
  )

  // 堆使用占进程申请内存的比例，用于进度条
  const heapPercent = computed(() => {
    const sys = data.value?.memory.sysMB ?? 0
    const heap = data.value?.memory.heapInuseMB ?? 0
    if (sys <= 0) return 0

    return Math.min(100, Math.round((heap / sys) * 100))
  })

  // 把秒数转成「x 天 x 小时 x 分」
  const formatUptime = (seconds?: number): string => {
    if (!seconds) return '-'

    const day = Math.floor(seconds / 86400)
    const hour = Math.floor((seconds % 86400) / 3600)
    const minute = Math.floor((seconds % 3600) / 60)

    if (day > 0) return `${day} 天 ${hour} 小时`
    if (hour > 0) return `${hour} 小时 ${minute} 分`

    return `${minute} 分`
  }

  // 探测失败时给面板加红边
  const healthClass = (ok?: boolean): string => (ok === false ? 'is-abnormal' : '')

  const loadOverview = async (): Promise<void> => {
    loading.value = true
    try {
      data.value = await fetchGetMonitorOverview()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  const stopTimer = (): void => {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
  }

  const startTimer = (): void => {
    stopTimer()
    timer = setInterval(loadOverview, 30_000)
  }

  watch(autoRefresh, (enabled) => (enabled ? startTimer() : stopTimer()))

  onMounted(async () => {
    await loadOverview()
    if (autoRefresh.value) startTimer()
  })

  onUnmounted(stopTimer)
</script>

<style lang="scss" scoped>
  .monitor-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .status-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    margin-right: 6px;
    vertical-align: middle;
    border-radius: 50%;

    &.is-ok {
      background: var(--el-color-success);
    }

    &.is-bad {
      background: var(--el-color-white);
    }
  }

  .monitor-card {
    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .monitor-grid {
    display: grid;
    gap: 16px;
    grid-template-columns: repeat(4, minmax(0, 1fr));

    @media (max-width: 1280px) {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    @media (max-width: 640px) {
      grid-template-columns: 1fr;
    }
  }

  .stat-panel {
    padding: 16px;
    background: var(--el-fill-color-lighter);
    border: 1px solid transparent;
    border-radius: 10px;

    &.is-abnormal {
      background: var(--el-color-danger-light-9);
      border-color: var(--el-color-danger-light-7);
    }

    &.is-table {
      background: transparent;
      border-color: var(--el-border-color-lighter);
    }
  }

  .span-2 {
    grid-column: span 2;

    @media (max-width: 640px) {
      grid-column: span 1;
    }
  }

  .panel-title {
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
  }

  .panel-value {
    font-size: 24px;
    font-weight: 600;
    line-height: 1.2;
    color: var(--el-text-color-primary);
  }

  .panel-sub {
    margin-top: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .panel-error {
    margin-top: 8px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-color-danger);
    word-break: break-all;
  }

  .kv-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px dashed var(--el-border-color-lighter);
  }

  .kv {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 12px;
    color: var(--el-text-color-secondary);

    b {
      font-weight: 500;
      color: var(--el-text-color-primary);
    }
  }

  .progress-row {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-top: 12px;
  }

  .progress-track {
    flex: 1;
    height: 6px;
    overflow: hidden;
    background: var(--el-fill-color-dark);
    border-radius: 3px;
  }

  .progress-fill {
    height: 100%;
    background: var(--el-color-primary);
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .metric-row {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    margin-top: 4px;
  }

  .metric {
    text-align: center;
  }

  .metric-value {
    font-size: 22px;
    font-weight: 600;
    color: var(--el-text-color-primary);

    &.is-danger {
      color: var(--el-color-danger);
    }
  }

  .metric-label {
    margin-top: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
</style>
