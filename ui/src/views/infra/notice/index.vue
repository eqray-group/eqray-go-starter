<!-- 基础设施 / 消息通知：渠道配置、模板管理、发送记录 -->
<template>
  <div class="infra-notice-page art-full-height">
    <ElCard class="notice-card art-table-card">
      <ElTabs v-model="activeTab" class="notice-tabs">
        <!-- 渠道 -->
        <ElTabPane name="channel">
          <template #label>
            <span>渠道</span>
            <span v-if="channelPagination.total" class="tab-count">
              {{ channelPagination.total }}
            </span>
          </template>

          <div v-if="activeTab === 'channel'" class="pane">
            <ArtSearchBar
              v-model="channelSearchForm"
              :items="channelSearchItems"
              :rules="{}"
              @reset="channelResetSearchParams"
              @search="handleChannelSearch"
            />

            <ArtTableHeader
              v-model:columns="channelColumnChecks"
              :loading="channelLoading"
              @refresh="channelRefresh"
            >
              <template #left>
                <ElSpace wrap>
                  <ElButton
                    v-auth="'infra:notice:update'"
                    type="primary"
                    @click="showChannelDialog()"
                    v-ripple
                  >
                    新增渠道
                  </ElButton>
                  <ElTag type="info" effect="plain">邮件走 SMTP，群消息走 Webhook</ElTag>
                </ElSpace>
              </template>
            </ArtTableHeader>

            <div class="pane-table">
              <ArtTable
                :show-table-header="false"
                :loading="channelLoading"
                :data="channelData"
                :columns="channelColumns"
                :pagination="channelPagination"
                @pagination:size-change="channelHandleSizeChange"
                @pagination:current-change="channelHandleCurrentChange"
              >
                <template #operation="{ row }">
                  <ElButton
                    v-auth="'infra:notice:send'"
                    link
                    type="primary"
                    @click="showTestDialog(row)"
                  >
                    测试
                  </ElButton>
                  <ElButton
                    v-auth="'infra:notice:update'"
                    link
                    type="primary"
                    @click="showChannelDialog(row)"
                  >
                    编辑
                  </ElButton>
                  <ElButton
                    v-auth="'infra:notice:delete'"
                    link
                    type="danger"
                    @click="handleDeleteChannel(row)"
                  >
                    删除
                  </ElButton>
                </template>
              </ArtTable>
            </div>
          </div>
        </ElTabPane>

        <!-- 模板 -->
        <ElTabPane name="template">
          <template #label>
            <span>模板</span>
            <span v-if="templatePagination.total" class="tab-count">
              {{ templatePagination.total }}
            </span>
          </template>

          <div v-if="activeTab === 'template'" class="pane">
            <ArtSearchBar
              v-model="templateSearchForm"
              :items="templateSearchItems"
              :rules="{}"
              @reset="templateResetSearchParams"
              @search="handleTemplateSearch"
            />

            <ArtTableHeader
              v-model:columns="templateColumnChecks"
              :loading="templateLoading"
              @refresh="templateRefresh"
            >
              <template #left>
                <ElSpace wrap>
                  <ElButton
                    v-auth="'infra:notice:update'"
                    type="primary"
                    @click="showTemplateDialog()"
                    v-ripple
                  >
                    新增模板
                  </ElButton>
                  <ElTag type="info" effect="plain">业务按编码调用，内容支持变量占位</ElTag>
                </ElSpace>
              </template>
            </ArtTableHeader>

            <div class="pane-table">
              <ArtTable
                :show-table-header="false"
                :loading="templateLoading"
                :data="templateData"
                :columns="templateColumns"
                :pagination="templatePagination"
                @pagination:size-change="templateHandleSizeChange"
                @pagination:current-change="templateHandleCurrentChange"
              >
                <template #operation="{ row }">
                  <ElButton
                    v-auth="'infra:notice:update'"
                    link
                    type="primary"
                    @click="showTemplateDialog(row)"
                  >
                    编辑
                  </ElButton>
                  <ElButton
                    v-auth="'infra:notice:delete'"
                    link
                    type="danger"
                    @click="handleDeleteTemplate(row)"
                  >
                    删除
                  </ElButton>
                </template>
              </ArtTable>
            </div>
          </div>
        </ElTabPane>

        <!-- 发送记录 -->
        <ElTabPane name="log">
          <template #label>
            <span>发送记录</span>
            <span v-if="logPagination.total" class="tab-count">{{ logPagination.total }}</span>
          </template>

          <div v-if="activeTab === 'log'" class="pane">
            <ArtSearchBar
              v-model="logSearchForm"
              :items="logSearchItems"
              :rules="{}"
              :sanitize-output="{ keepZero: true }"
              @reset="logResetSearchParams"
              @search="handleLogSearch"
            />

            <ArtTableHeader
              v-model:columns="logColumnChecks"
              :loading="logLoading"
              @refresh="logRefresh"
            >
              <template #left>
                <ElTag type="info" effect="plain">展开行可看失败原因与完整正文</ElTag>
              </template>
            </ArtTableHeader>

            <div class="pane-table">
              <ArtTable
                :show-table-header="false"
                :loading="logLoading"
                :data="logData"
                :columns="logColumns"
                :pagination="logPagination"
                @pagination:size-change="logHandleSizeChange"
                @pagination:current-change="logHandleCurrentChange"
              />
            </div>
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <NoticeChannelDialog
      v-model:visible="channelDialogVisible"
      :channel-data="currentChannel"
      :types="types"
      @success="channelRefresh"
    />

    <NoticeTemplateDialog
      v-model:visible="templateDialogVisible"
      :template-data="currentTemplate"
      :types="types"
      @success="templateRefresh"
    />

    <NoticeTestDialog
      v-model:visible="testDialogVisible"
      :channel="testChannel"
      :type-name="testChannel ? typeName(testChannel.type) : ''"
      @success="logRefresh"
    />
  </div>
</template>

<script setup lang="ts">
  import { h } from 'vue'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchDeleteNoticeChannel,
    fetchDeleteNoticeTemplate,
    fetchGetNoticeChannels,
    fetchGetNoticeLogs,
    fetchGetNoticeTemplates,
    fetchGetNoticeTypes
  } from '@/api/infra/notice'
  import NoticeChannelDialog from './modules/notice-channel-dialog.vue'
  import NoticeLogDetail from './modules/notice-log-detail.vue'
  import NoticeTemplateDialog from './modules/notice-template-dialog.vue'
  import NoticeTestDialog from './modules/notice-test-dialog.vue'

  defineOptions({ name: 'InfraNotice' })

  type ChannelItem = Api.Infra.NoticeChannelItem
  type TemplateItem = Api.Infra.NoticeTemplateItem
  type LogItem = Api.Infra.NoticeLogItem
  type TypeSchema = Api.Infra.NoticeTypeSchema

  const activeTab = ref<'channel' | 'template' | 'log'>('channel')
  const types = ref<TypeSchema[]>([])

  const typeOptions = computed(() =>
    types.value.map((item) => ({ label: item.name, value: item.type }))
  )

  const typeName = (type: string): string =>
    types.value.find((item) => item.type === type)?.name || type

  const typeTag = (type: string): 'primary' | 'success' | 'warning' | 'info' => {
    switch (type) {
      case 'email':
        return 'primary'
      case 'wecom':
        return 'success'
      case 'dingtalk':
        return 'warning'
      default:
        return 'info'
    }
  }

  // 渠道配置是 JSON 文本，列表里只给一句人话摘要，避免把密码等字段直接铺在表格里
  const configSummary = (row: ChannelItem): string => {
    try {
      const config = JSON.parse(row.config || '{}') as Record<string, string>
      if (row.type === 'email') {
        if (!config.host) return '未配置'
        return `${config.host}:${config.port || 25}`
      }
      if (!config.webhook) return '未配置'
      return config.secret ? 'Webhook + 签名' : 'Webhook 已配置'
    } catch {
      return '配置格式异常'
    }
  }

  // ---------- 渠道 ----------
  const channelSearchForm = ref<Api.Infra.NoticeSearchParams>({})
  const channelSearchItems = computed(() => [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      placeholder: '渠道名称',
      clearable: true
    },
    {
      label: '类型',
      key: 'type',
      type: 'select',
      props: { placeholder: '全部类型', clearable: true, options: typeOptions.value }
    }
  ])

  const {
    columns: channelColumns,
    columnChecks: channelColumnChecks,
    data: channelData,
    loading: channelLoading,
    pagination: channelPagination,
    searchParams: channelSearchParams,
    getData: channelGetData,
    resetSearchParams: channelResetSearchParams,
    handleSizeChange: channelHandleSizeChange,
    handleCurrentChange: channelHandleCurrentChange,
    refreshData: channelRefresh
  } = useTable({
    core: {
      apiFn: fetchGetNoticeChannels,
      apiParams: { pageNo: 1, pageSize: 10 },
      columnsFactory: () => [
        { prop: 'name', label: '渠道名称', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'type',
          label: '类型',
          width: 150,
          formatter: (row: ChannelItem) =>
            h(ElTag, { type: typeTag(row.type), size: 'small', effect: 'plain' }, () =>
              typeName(row.type)
            )
        },
        {
          prop: 'config',
          label: '连接',
          minWidth: 170,
          formatter: (row: ChannelItem) => configSummary(row)
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: ChannelItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info', size: 'small' }, () =>
              row.status === 1 ? '启用' : '停用'
            )
        },
        { prop: 'remark', label: '备注', minWidth: 150, showOverflowTooltip: true },
        { prop: 'updatedAt', label: '更新时间', width: 165 },
        { prop: 'operation', label: '操作', width: 160, fixed: 'right', useSlot: true }
      ]
    }
  })

  const handleChannelSearch = (params: Api.Infra.NoticeSearchParams): void => {
    Object.assign(channelSearchForm.value, params)
    Object.assign(channelSearchParams, { keyword: undefined, type: undefined }, params)
    channelGetData()
  }

  // ---------- 模板 ----------
  const templateSearchForm = ref<Api.Infra.NoticeSearchParams>({})
  const templateSearchItems = computed(() => [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      placeholder: '编码 / 名称',
      clearable: true
    },
    {
      label: '类型',
      key: 'type',
      type: 'select',
      props: { placeholder: '全部类型', clearable: true, options: typeOptions.value }
    }
  ])

  const {
    columns: templateColumns,
    columnChecks: templateColumnChecks,
    data: templateData,
    loading: templateLoading,
    pagination: templatePagination,
    searchParams: templateSearchParams,
    getData: templateGetData,
    resetSearchParams: templateResetSearchParams,
    handleSizeChange: templateHandleSizeChange,
    handleCurrentChange: templateHandleCurrentChange,
    refreshData: templateRefresh
  } = useTable({
    core: {
      apiFn: fetchGetNoticeTemplates,
      apiParams: { pageNo: 1, pageSize: 10 },
      columnsFactory: () => [
        { prop: 'code', label: '编码', width: 170, showOverflowTooltip: true },
        { prop: 'name', label: '名称', minWidth: 150, showOverflowTooltip: true },
        {
          prop: 'type',
          label: '类型',
          width: 150,
          formatter: (row: TemplateItem) =>
            h(ElTag, { type: typeTag(row.type), size: 'small', effect: 'plain' }, () =>
              typeName(row.type)
            )
        },
        { prop: 'title', label: '标题', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: TemplateItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info', size: 'small' }, () =>
              row.status === 1 ? '启用' : '停用'
            )
        },
        { prop: 'operation', label: '操作', width: 130, fixed: 'right', useSlot: true }
      ]
    }
  })

  const handleTemplateSearch = (params: Api.Infra.NoticeSearchParams): void => {
    Object.assign(templateSearchForm.value, params)
    Object.assign(templateSearchParams, { keyword: undefined, type: undefined }, params)
    templateGetData()
  }

  // ---------- 发送记录 ----------
  const logSearchForm = ref<Api.Infra.NoticeSearchParams>({})
  const logSearchItems = computed(() => [
    {
      label: '类型',
      key: 'type',
      type: 'select',
      props: { placeholder: '全部类型', clearable: true, options: typeOptions.value }
    },
    {
      label: '结果',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部结果',
        clearable: true,
        options: [
          { label: '成功', value: 1 },
          { label: '失败', value: 0 }
        ]
      }
    }
  ])

  const {
    columns: logColumns,
    columnChecks: logColumnChecks,
    data: logData,
    loading: logLoading,
    pagination: logPagination,
    searchParams: logSearchParams,
    getData: logGetData,
    resetSearchParams: logResetSearchParams,
    handleSizeChange: logHandleSizeChange,
    handleCurrentChange: logHandleCurrentChange,
    refreshData: logRefresh
  } = useTable({
    core: {
      apiFn: fetchGetNoticeLogs,
      apiParams: { pageNo: 1, pageSize: 20 },
      columnsFactory: () => [
        { type: 'expand', formatter: (row: LogItem) => h(NoticeLogDetail, { row }) },
        { prop: 'createdAt', label: '时间', width: 165 },
        { prop: 'channelName', label: '渠道', width: 150, showOverflowTooltip: true },
        {
          prop: 'type',
          label: '类型',
          width: 140,
          formatter: (row: LogItem) =>
            h(ElTag, { type: typeTag(row.type), size: 'small', effect: 'plain' }, () =>
              typeName(row.type)
            )
        },
        { prop: 'templateCode', label: '模板', width: 150, showOverflowTooltip: true },
        { prop: 'target', label: '接收目标', minWidth: 150, showOverflowTooltip: true },
        {
          prop: 'status',
          label: '结果',
          width: 90,
          formatter: (row: LogItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' }, () =>
              row.status === 1 ? '成功' : '失败'
            )
        },
        {
          prop: 'duration',
          label: '耗时',
          width: 90,
          formatter: (row: LogItem) => `${row.duration} ms`
        }
      ]
    }
  })

  const handleLogSearch = (params: Api.Infra.NoticeSearchParams): void => {
    Object.assign(logSearchForm.value, params)
    Object.assign(logSearchParams, { type: undefined, status: undefined }, params)
    logGetData()
  }

  // ---------- 弹窗 ----------
  const channelDialogVisible = ref(false)
  const currentChannel = ref<ChannelItem | null>(null)
  const templateDialogVisible = ref(false)
  const currentTemplate = ref<TemplateItem | null>(null)
  const testDialogVisible = ref(false)
  const testChannel = ref<ChannelItem | null>(null)

  const showChannelDialog = (row?: ChannelItem): void => {
    currentChannel.value = row ?? null
    channelDialogVisible.value = true
  }

  const showTemplateDialog = (row?: TemplateItem): void => {
    currentTemplate.value = row ?? null
    templateDialogVisible.value = true
  }

  const showTestDialog = (row: ChannelItem): void => {
    testChannel.value = row
    testDialogVisible.value = true
  }

  // ---------- 删除 ----------
  const handleDeleteChannel = async (row: ChannelItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`删除渠道「${row.name}」？已发送的记录会保留。`, '删除渠道', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteNoticeChannel(row.id)
      ElMessage.success('已删除')
      channelRefresh()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleDeleteTemplate = async (row: TemplateItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `删除模板「${row.name}」？业务再按编码调用会报「模板不存在」。`,
        '删除模板',
        {
          confirmButtonText: '删除',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }

    try {
      await fetchDeleteNoticeTemplate(row.id)
      ElMessage.success('已删除')
      templateRefresh()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(async () => {
    try {
      types.value = (await fetchGetNoticeTypes()) ?? []
    } catch {
      types.value = []
    }
  })
</script>

<style lang="scss" scoped>
  .notice-card {
    display: flex;
    flex: 1;
    min-height: 0;

    :deep(.el-card__body) {
      display: flex;
      flex: 1;
      flex-direction: column;
      min-height: 0;
      padding: 0 16px 12px;
    }
  }

  .notice-tabs {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;

    :deep(.el-tabs__header) {
      margin-bottom: 10px;
    }

    :deep(.el-tabs__content) {
      flex: 1;
      min-height: 0;
    }

    :deep(.el-tab-pane) {
      height: 100%;
    }
  }

  .tab-count {
    padding: 0 5px;
    margin-left: 4px;
    font-size: 11px;
    line-height: 16px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color);
    border-radius: 8px;
  }

  // 每个 tab 内部：搜索栏 + 工具条固定，表格占满剩余高度
  .pane {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .pane-table {
    flex: 1;
    min-height: 0;
  }
</style>
