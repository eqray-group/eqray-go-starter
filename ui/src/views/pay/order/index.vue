<!-- 支付订单页面：收单记录、状态同步与关单 -->
<template>
  <div class="pay-order-page art-full-height">
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

    <PayOrderDetailDialog v-model:visible="detailVisible" :order-data="currentOrder" />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchClosePayOrder, fetchGetPayOrderList, fetchSyncPayOrder } from '@/api/pay/order'
  import { ElCard, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import PayOrderDetailDialog from './modules/pay-order-detail-dialog.vue'

  defineOptions({ name: 'PayOrder' })

  type OrderItem = Api.Pay.OrderItem

  // 订单状态：与后端 PayOrderStatus 对齐
  const STATUS_MAP: Record<
    number,
    { label: string; type: 'info' | 'warning' | 'success' | 'danger' }
  > = {
    0: { label: '待支付', type: 'info' },
    10: { label: '支付中', type: 'warning' },
    20: { label: '已支付', type: 'success' },
    30: { label: '已关闭', type: 'danger' }
  }

  // 退款进度：与后端 PayOrderRefundStatus 对齐
  const REFUND_MAP: Record<number, string> = {
    0: '未退款',
    10: '部分退款',
    20: '全额退款'
  }

  // 金额（分）转展示用「元」
  const formatPrice = (price: number): string => `${((price || 0) / 100).toFixed(2)} 元`

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Pay.OrderSearchParams>({
    merchantOrderNo: undefined,
    bizOrderNo: undefined,
    channelCode: undefined,
    status: undefined,
    createTimeRange: null
  })

  // 渠道编码选项：与后端 channel.Metas 的 9 种渠道一致
  const channelOptions = [
    { label: '微信扫码', value: 'wx_native' },
    { label: '微信 H5', value: 'wx_h5' },
    { label: '微信 JSAPI', value: 'wx_jsapi' },
    { label: '微信小程序', value: 'wx_mini' },
    { label: '微信 APP', value: 'wx_app' },
    { label: '支付宝电脑网站', value: 'ali_pc' },
    { label: '支付宝手机网站', value: 'ali_wap' },
    { label: '支付宝 APP', value: 'ali_app' },
    { label: '支付宝当面付', value: 'ali_qr' }
  ]

  const searchItems = [
    {
      label: '支付单号',
      key: 'merchantOrderNo',
      type: 'input',
      placeholder: '请输入支付单号',
      clearable: true
    },
    {
      label: '业务单号',
      key: 'bizOrderNo',
      type: 'input',
      placeholder: '请输入业务单号',
      clearable: true
    },
    {
      label: '支付渠道',
      key: 'channelCode',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: channelOptions
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '待支付', value: 0 },
          { label: '支付中', value: 10 },
          { label: '已支付', value: 20 },
          { label: '已关闭', value: 30 }
        ]
      }
    },
    {
      label: '创建时间',
      key: 'createTimeRange',
      type: 'datetimerange',
      span: 24,
      props: {
        type: 'datetimerange',
        valueFormat: 'YYYY-MM-DD HH:mm:ss',
        startPlaceholder: '开始时间',
        endPlaceholder: '结束时间',
        rangeSeparator: '至',
        style: 'width: 100%'
      }
    }
  ]

  const detailVisible = ref(false)
  const currentOrder = ref<OrderItem | null>(null)

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
    refreshData,
    refreshUpdate
  } = useTable({
    core: {
      apiFn: fetchGetPayOrderList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'merchantOrderNo',
          label: '支付单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'bizOrderNo',
          label: '业务单',
          minWidth: 190,
          formatter: (row: OrderItem) => `${row.bizType} / ${row.bizOrderNo}`
        },
        {
          prop: 'channelCode',
          label: '渠道',
          width: 130,
          formatter: (row: OrderItem) => h(ElTag, { type: 'info' }, () => row.channelCode)
        },
        {
          prop: 'price',
          label: '金额',
          width: 120,
          formatter: (row: OrderItem) => formatPrice(row.price)
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: OrderItem) => {
            const status = STATUS_MAP[row.status] ?? { label: '未知', type: 'info' as const }
            return h(ElTag, { type: status.type }, () => status.label)
          }
        },
        {
          prop: 'refundStatus',
          label: '退款',
          width: 100,
          formatter: (row: OrderItem) => REFUND_MAP[row.refundStatus] ?? '-'
        },
        { prop: 'successTime', label: '支付时间', width: 170 },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 250,
          fixed: 'right',
          formatter: (row: OrderItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'action',
                icon: 'ri:eye-line',
                text: '详情',
                onClick: () => showDetail(row)
              }),
              ...(hasAuth('pay:order:update') && (row.status === 0 || row.status === 10)
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      icon: 'ri:refresh-line',
                      text: '同步',
                      onClick: () => handleSync(row)
                    }),
                    h(ArtButtonTable, {
                      type: 'action',
                      icon: 'ri:close-circle-line',
                      text: '关闭',
                      onClick: () => handleClose(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索栏的日期范围拆成后端认识的 beginTime / endTime，避免把数组直接发给接口
  const handleSearch = (params: Api.Pay.OrderSearchParams) => {
    const { createTimeRange, ...rest } = params
    const [beginTime, endTime] = createTimeRange ?? []

    Object.assign(searchParams, rest, {
      beginTime: beginTime || undefined,
      endTime: endTime || undefined
    })
    getData()
  }

  const showDetail = (row: OrderItem): void => {
    currentOrder.value = row
    detailVisible.value = true
  }

  // 主动查单：兜底「渠道已收款但回调未达」，同步后本地状态可能变为已支付
  const handleSync = async (row: OrderItem): Promise<void> => {
    try {
      await fetchSyncPayOrder(row.id)
      ElMessage.success('已向渠道同步最新状态')
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleClose = async (row: OrderItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定关闭订单「${row.merchantOrderNo}」吗？关闭后会同时关闭渠道侧订单。`,
        '关闭支付订单',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchClosePayOrder(row.id)
      ElMessage.success('订单已关闭')
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
