<!-- 支付回调日志页面：排查「为什么没到账」时主要看这张表 -->
<template>
  <div class="pay-notify-log-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog v-model="detailVisible" title="回调报文" width="820px" align-center>
      <ElDescriptions :column="2" border class="mb-4">
        <ElDescriptionsItem label="渠道">{{ currentLog?.channelCode || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="支付单号">
          {{ currentLog?.merchantOrderNo || '-' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem label="验签结果">
          <ElTag :type="currentLog?.verifyStatus === 1 ? 'success' : 'danger'">
            {{ currentLog?.verifyStatus === 1 ? '通过' : '失败' }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="处理结果">
          <ElTag :type="currentLog?.handleStatus === 1 ? 'success' : 'danger'">
            {{ currentLog?.handleStatus === 1 ? '成功' : '失败' }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem v-if="currentLog?.errorMsg" label="失败原因" :span="2">
          {{ currentLog?.errorMsg }}
        </ElDescriptionsItem>
      </ElDescriptions>

      <ElTabs v-model="activeTab">
        <ElTabPane label="解密后报文" name="decrypted">
          <pre class="notify-body">{{ prettyBody(currentLog?.decryptedBody) }}</pre>
        </ElTabPane>
        <ElTabPane label="原始报文" name="raw">
          <pre class="notify-body">{{ prettyBody(currentLog?.requestBody) }}</pre>
        </ElTabPane>
      </ElTabs>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayNotifyLogList } from '@/api/pay/notify-log'
  import { ElCard, ElDialog, ElTabPane, ElTabs, ElTag } from 'element-plus'

  defineOptions({ name: 'PayNotifyLog' })

  type NotifyLogItem = Api.Pay.NotifyLogItem

  const searchForm = ref<Api.Pay.NotifyLogSearchParams>({
    channelCode: undefined,
    type: undefined,
    merchantOrderNo: undefined,
    handleStatus: undefined
  })

  const searchItems = [
    {
      label: '渠道编码',
      key: 'channelCode',
      type: 'input',
      placeholder: '如 wx_mini',
      clearable: true
    },
    {
      label: '通知类型',
      key: 'type',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '支付通知', value: 'pay' },
          { label: '退款通知', value: 'refund' }
        ]
      }
    },
    {
      label: '支付单号',
      key: 'merchantOrderNo',
      type: 'input',
      placeholder: '请输入支付单号',
      clearable: true
    },
    {
      label: '处理结果',
      key: 'handleStatus',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '成功', value: 1 },
          { label: '失败', value: 0 }
        ]
      }
    }
  ]

  const detailVisible = ref(false)
  const currentLog = ref<NotifyLogItem | null>(null)
  const activeTab = ref('decrypted')

  // 报文尽量格式化展示，非 JSON（如表单）原样输出
  const prettyBody = (body?: string): string => {
    if (!body) return '（空）'
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetPayNotifyLogList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'channelCode', label: '渠道编码', width: 130 },
        {
          prop: 'type',
          label: '类型',
          width: 90,
          formatter: (row: NotifyLogItem) => (row.type === 'refund' ? '退款' : '支付')
        },
        {
          prop: 'merchantOrderNo',
          label: '支付单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'verifyStatus',
          label: '验签',
          width: 90,
          formatter: (row: NotifyLogItem) =>
            h(ElTag, { type: row.verifyStatus === 1 ? 'success' : 'danger' }, () =>
              row.verifyStatus === 1 ? '通过' : '失败'
            )
        },
        {
          prop: 'handleStatus',
          label: '处理',
          width: 90,
          formatter: (row: NotifyLogItem) =>
            h(ElTag, { type: row.handleStatus === 1 ? 'success' : 'danger' }, () =>
              row.handleStatus === 1 ? '成功' : '失败'
            )
        },
        {
          prop: 'errorMsg',
          label: '失败原因',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: NotifyLogItem) => row.errorMsg || '-'
        },
        { prop: 'createdAt', label: '回调时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row: NotifyLogItem) =>
            h(ArtButtonTable, {
              type: 'action',
              icon: 'ri:file-text-line',
              text: '报文',
              onClick: () => showDetail(row)
            })
        }
      ]
    }
  })

  const handleSearch = (params: Api.Pay.NotifyLogSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDetail = (row: NotifyLogItem): void => {
    currentLog.value = row
    activeTab.value = 'decrypted'
    detailVisible.value = true
  }
</script>

<style lang="scss" scoped>
  .notify-body {
    max-height: 420px;
    overflow: auto;
    padding: 12px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
