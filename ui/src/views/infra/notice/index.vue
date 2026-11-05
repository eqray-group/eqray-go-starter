<!-- 基础设施 / 消息通知：渠道配置、模板管理与发送记录 -->
<template>
  <div class="infra-notice-page art-full-height">
    <ElCard class="notice-card">
      <ElTabs v-model="activeTab" class="notice-tabs">
        <!-- 渠道 -->
        <ElTabPane label="渠道" name="channel">
          <div class="toolbar">
            <ElButton
              v-auth="'infra:notice:update'"
              type="primary"
              @click="showChannelDialog()"
              v-ripple
            >
              新增渠道
            </ElButton>
            <ElInput v-model="channelQuery.keyword" placeholder="渠道名称" clearable class="w-44" />
            <ElSelect v-model="channelQuery.type" placeholder="全部类型" clearable class="w-40">
              <ElOption
                v-for="item in types"
                :key="item.type"
                :label="item.name"
                :value="item.type"
              />
            </ElSelect>
            <ElButton @click="loadChannels(1)">查询</ElButton>
          </div>

          <ElTable v-loading="channelLoading" :data="channels" border>
            <ElTableColumn prop="name" label="渠道名称" min-width="160" show-overflow-tooltip />
            <ElTableColumn label="类型" width="150">
              <template #default="{ row }">{{ typeName(row.type) }}</template>
            </ElTableColumn>
            <ElTableColumn label="状态" width="90">
              <template #default="{ row }">
                <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="remark" label="备注" min-width="160" show-overflow-tooltip />
            <ElTableColumn prop="updatedAt" label="更新时间" width="170" />
            <ElTableColumn label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <ElButton v-auth="'infra:notice:send'" link type="primary" @click="handleTest(row)">
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
            </ElTableColumn>
            <template #empty>
              <ElEmpty description="还没有渠道，先新增一个邮件或群机器人" :image-size="90" />
            </template>
          </ElTable>

          <div class="pager">
            <ElPagination
              v-model:current-page="channelQuery.pageNo"
              v-model:page-size="channelQuery.pageSize"
              :total="channelTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              @current-change="loadChannels()"
              @size-change="loadChannels(1)"
            />
          </div>
        </ElTabPane>

        <!-- 模板 -->
        <ElTabPane label="模板" name="template">
          <div class="toolbar">
            <ElButton
              v-auth="'infra:notice:update'"
              type="primary"
              @click="showTemplateDialog()"
              v-ripple
            >
              新增模板
            </ElButton>
            <ElInput
              v-model="templateQuery.keyword"
              placeholder="编码 / 名称"
              clearable
              class="w-44"
            />
            <ElSelect v-model="templateQuery.type" placeholder="全部类型" clearable class="w-40">
              <ElOption
                v-for="item in types"
                :key="item.type"
                :label="item.name"
                :value="item.type"
              />
            </ElSelect>
            <ElButton @click="loadTemplates(1)">查询</ElButton>
          </div>

          <ElTable v-loading="templateLoading" :data="templates" border>
            <ElTableColumn prop="code" label="编码" width="180" show-overflow-tooltip />
            <ElTableColumn prop="name" label="名称" min-width="150" show-overflow-tooltip />
            <ElTableColumn label="类型" width="150">
              <template #default="{ row }">{{ typeName(row.type) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="title" label="标题" min-width="160" show-overflow-tooltip />
            <ElTableColumn label="状态" width="90">
              <template #default="{ row }">
                <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="操作" width="140" fixed="right">
              <template #default="{ row }">
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
            </ElTableColumn>
            <template #empty>
              <ElEmpty description="还没有模板，业务按编码调用" :image-size="90" />
            </template>
          </ElTable>

          <div class="pager">
            <ElPagination
              v-model:current-page="templateQuery.pageNo"
              v-model:page-size="templateQuery.pageSize"
              :total="templateTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              @current-change="loadTemplates()"
              @size-change="loadTemplates(1)"
            />
          </div>
        </ElTabPane>

        <!-- 发送记录 -->
        <ElTabPane label="发送记录" name="log">
          <div class="toolbar">
            <ElSelect v-model="logQuery.type" placeholder="全部类型" clearable class="w-40">
              <ElOption
                v-for="item in types"
                :key="item.type"
                :label="item.name"
                :value="item.type"
              />
            </ElSelect>
            <ElSelect v-model="logQuery.status" placeholder="全部结果" clearable class="w-36">
              <ElOption label="成功" :value="1" />
              <ElOption label="失败" :value="0" />
            </ElSelect>
            <ElButton @click="loadLogs(1)">查询</ElButton>
            <span class="text-xs text-g-600">失败原因与正文可展开查看</span>
          </div>

          <ElTable v-loading="logLoading" :data="logs" border>
            <ElTableColumn type="expand">
              <template #default="{ row }">
                <div class="log-detail">
                  <div v-if="row.error" class="mb-2">
                    <div class="log-label">失败原因</div>
                    <pre class="log-pre is-error">{{ row.error }}</pre>
                  </div>
                  <div class="log-label">发送内容</div>
                  <pre class="log-pre">{{ row.content || '（无内容）' }}</pre>
                </div>
              </template>
            </ElTableColumn>
            <ElTableColumn prop="createdAt" label="时间" width="170" />
            <ElTableColumn prop="channelName" label="渠道" width="150" show-overflow-tooltip />
            <ElTableColumn label="类型" width="140">
              <template #default="{ row }">{{ typeName(row.type) }}</template>
            </ElTableColumn>
            <ElTableColumn prop="templateCode" label="模板" width="150" show-overflow-tooltip />
            <ElTableColumn prop="target" label="接收目标" min-width="140" show-overflow-tooltip />
            <ElTableColumn label="结果" width="90">
              <template #default="{ row }">
                <ElTag :type="row.status === 1 ? 'success' : 'danger'" size="small">
                  {{ row.status === 1 ? '成功' : '失败' }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="耗时" width="90">
              <template #default="{ row }">{{ row.duration }} ms</template>
            </ElTableColumn>
            <template #empty>
              <ElEmpty description="还没有发送记录" :image-size="90" />
            </template>
          </ElTable>

          <div class="pager">
            <ElPagination
              v-model:current-page="logQuery.pageNo"
              v-model:page-size="logQuery.pageSize"
              :total="logTotal"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @current-change="loadLogs()"
              @size-change="loadLogs(1)"
            />
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <NoticeChannelDialog
      v-model:visible="channelDialogVisible"
      :channel-data="currentChannel"
      :types="types"
      @success="loadChannels()"
    />

    <NoticeTemplateDialog
      v-model:visible="templateDialogVisible"
      :template-data="currentTemplate"
      :types="types"
      @success="loadTemplates()"
    />
  </div>
</template>

<script setup lang="ts">
  import {
    ElButton,
    ElCard,
    ElEmpty,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElPagination,
    ElSelect,
    ElTabPane,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTag
  } from 'element-plus'
  import {
    fetchDeleteNoticeChannel,
    fetchDeleteNoticeTemplate,
    fetchGetNoticeChannels,
    fetchGetNoticeLogs,
    fetchGetNoticeTemplates,
    fetchGetNoticeTypes,
    fetchTestNoticeChannel
  } from '@/api/infra/notice'
  import NoticeChannelDialog from './modules/notice-channel-dialog.vue'
  import NoticeTemplateDialog from './modules/notice-template-dialog.vue'

  defineOptions({ name: 'InfraNotice' })

  type ChannelItem = Api.Infra.NoticeChannelItem
  type TemplateItem = Api.Infra.NoticeTemplateItem
  type LogItem = Api.Infra.NoticeLogItem
  type TypeSchema = Api.Infra.NoticeTypeSchema

  const activeTab = ref<'channel' | 'template' | 'log'>('channel')
  const types = ref<TypeSchema[]>([])

  // ---------- 渠道 ----------
  const channels = ref<ChannelItem[]>([])
  const channelTotal = ref(0)
  const channelLoading = ref(false)
  const channelDialogVisible = ref(false)
  const currentChannel = ref<ChannelItem | null>(null)
  const channelQuery = reactive({
    keyword: '',
    type: '',
    pageNo: 1,
    pageSize: 10
  })

  // ---------- 模板 ----------
  const templates = ref<TemplateItem[]>([])
  const templateTotal = ref(0)
  const templateLoading = ref(false)
  const templateDialogVisible = ref(false)
  const currentTemplate = ref<TemplateItem | null>(null)
  const templateQuery = reactive({
    keyword: '',
    type: '',
    pageNo: 1,
    pageSize: 10
  })

  // ---------- 发送记录 ----------
  const logs = ref<LogItem[]>([])
  const logTotal = ref(0)
  const logLoading = ref(false)
  const logQuery = reactive({
    type: '',
    status: undefined as number | undefined,
    pageNo: 1,
    pageSize: 20
  })

  const typeName = (type: string): string =>
    types.value.find((item) => item.type === type)?.name || type

  const loadTypes = async (): Promise<void> => {
    try {
      types.value = (await fetchGetNoticeTypes()) ?? []
    } catch {
      types.value = []
    }
  }

  const loadChannels = async (page?: number): Promise<void> => {
    if (page) channelQuery.pageNo = page

    channelLoading.value = true
    try {
      const res = await fetchGetNoticeChannels({
        keyword: channelQuery.keyword || undefined,
        type: channelQuery.type || undefined,
        pageNo: channelQuery.pageNo,
        pageSize: channelQuery.pageSize
      })
      channels.value = res?.list ?? []
      channelTotal.value = res?.total ?? 0
    } catch {
      channels.value = []
      channelTotal.value = 0
    } finally {
      channelLoading.value = false
    }
  }

  const loadTemplates = async (page?: number): Promise<void> => {
    if (page) templateQuery.pageNo = page

    templateLoading.value = true
    try {
      const res = await fetchGetNoticeTemplates({
        keyword: templateQuery.keyword || undefined,
        type: templateQuery.type || undefined,
        pageNo: templateQuery.pageNo,
        pageSize: templateQuery.pageSize
      })
      templates.value = res?.list ?? []
      templateTotal.value = res?.total ?? 0
    } catch {
      templates.value = []
      templateTotal.value = 0
    } finally {
      templateLoading.value = false
    }
  }

  const loadLogs = async (page?: number): Promise<void> => {
    if (page) logQuery.pageNo = page

    logLoading.value = true
    try {
      const res = await fetchGetNoticeLogs({
        type: logQuery.type || undefined,
        status: logQuery.status,
        pageNo: logQuery.pageNo,
        pageSize: logQuery.pageSize
      })
      logs.value = res?.list ?? []
      logTotal.value = res?.total ?? 0
    } catch {
      logs.value = []
      logTotal.value = 0
    } finally {
      logLoading.value = false
    }
  }

  const showChannelDialog = (row?: ChannelItem): void => {
    currentChannel.value = row ?? null
    channelDialogVisible.value = true
  }

  const showTemplateDialog = (row?: TemplateItem): void => {
    currentTemplate.value = row ?? null
    templateDialogVisible.value = true
  }

  // 测试发送：邮件渠道先问收件人，机器人渠道直接确认
  const handleTest = async (row: ChannelItem): Promise<void> => {
    let target = ''

    if (row.type === 'email') {
      try {
        const result = await ElMessageBox.prompt('发送到哪个邮箱？', '测试发送', {
          confirmButtonText: '发送',
          cancelButtonText: '取消',
          inputPlaceholder: 'name@example.com',
          inputValue: ''
        })
        target = (result as { value: string }).value?.trim() || ''
      } catch {
        return
      }
      if (!target) {
        ElMessage.warning('请填写收件人')
        return
      }
    } else {
      try {
        await ElMessageBox.confirm(`向「${row.name}」发送一条测试消息？`, '测试发送', {
          confirmButtonText: '发送',
          cancelButtonText: '取消',
          type: 'warning'
        })
      } catch {
        return
      }
    }

    try {
      const log = await fetchTestNoticeChannel(row.id, target || undefined)
      if (log?.status === 1) {
        ElMessage.success(`发送成功，耗时 ${log.duration} ms`)
      } else {
        ElMessage.error(`发送失败：${log?.error || '未知原因'}`)
      }
      loadLogs(1)
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

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
      loadChannels()
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
      loadTemplates()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(async () => {
    await loadTypes()
    await Promise.all([loadChannels(), loadTemplates(), loadLogs()])
  })
</script>

<style lang="scss" scoped>
  .notice-card {
    :deep(.el-card__body) {
      padding: 8px 20px 20px;
    }
  }

  .notice-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 16px;
    }
  }

  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-bottom: 12px;

    .w-44 {
      width: 176px;
    }

    .w-40 {
      width: 160px;
    }

    .w-36 {
      width: 144px;
    }
  }

  .pager {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }

  .log-detail {
    padding: 8px 16px;
  }

  .log-label {
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .log-pre {
    max-height: 220px;
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
