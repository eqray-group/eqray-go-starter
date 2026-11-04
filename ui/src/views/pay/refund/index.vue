<!-- 退款订单页面：退款记录查询与后台直接发起退款 -->
<template>
  <div class="pay-refund-page art-full-height">
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
            <ElButton v-auth="'pay:refund:create'" @click="dialogVisible = true" v-ripple>
              发起退款
            </ElButton>
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

    <PayRefundCreateDialog v-model:visible="dialogVisible" @success="refreshCreate" />
    <PayRefundDetailDialog v-model:visible="detailVisible" :refund-data="currentRefund" />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayRefundList } from '@/api/pay/refund'
  import { ElButton, ElCard, ElSpace, ElTag } from 'element-plus'
  import PayRefundCreateDialog from './modules/pay-refund-create-dialog.vue'
  import PayRefundDetailDialog from './modules/pay-refund-detail-dialog.vue'

  defineOptions({ name: 'PayRefund' })

  type RefundItem = Api.Pay.RefundItem

  // 退款状态：与后端 PayRefundStatus 对齐
  const STATUS_MAP: Record<
    number,
    { label: string; type: 'info' | 'warning' | 'success' | 'danger' }
  > = {
    0: { label: '待退款', type: 'info' },
    10: { label: '退款中', type: 'warning' },
    20: { label: '已退款', type: 'success' },
    30: { label: '退款失败', type: 'danger' }
  }

  // 金额（分）转展示用「元」
  const formatPrice = (price: number): string => `${((price || 0) / 100).toFixed(2)} 元`

  const searchForm = ref<Api.Pay.RefundSearchParams>({
    merchantRefundNo: undefined,
    merchantOrderNo: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '退款单号',
      key: 'merchantRefundNo',
      type: 'input',
      placeholder: '请输入退款单号',
      clearable: true
    },
    {
      label: '支付单号',
      key: 'merchantOrderNo',
      type: 'input',
      placeholder: '请输入支付单号',
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
          { label: '待退款', value: 0 },
          { label: '退款中', value: 10 },
          { label: '已退款', value: 20 },
          { label: '退款失败', value: 30 }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const detailVisible = ref(false)
  const currentRefund = ref<RefundItem | null>(null)

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
    refreshCreate
  } = useTable({
    core: {
      apiFn: fetchGetPayRefundList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'merchantRefundNo',
          label: '退款单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'merchantOrderNo',
          label: '支付单号',
          minWidth: 190,
          showOverflowTooltip: true
        },
        {
          prop: 'refundPrice',
          label: '退款金额',
          width: 120,
          formatter: (row: RefundItem) => formatPrice(row.refundPrice)
        },
        {
          prop: 'status',
          label: '状态',
          width: 110,
          formatter: (row: RefundItem) => {
            const status = STATUS_MAP[row.status] ?? { label: '未知', type: 'info' as const }
            return h(ElTag, { type: status.type }, () => status.label)
          }
        },
        { prop: 'reason', label: '退款原因', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'channelRefundNo',
          label: '渠道退款单号',
          minWidth: 170,
          showOverflowTooltip: true
        },
        {
          prop: 'errorMsg',
          label: '失败原因',
          minWidth: 180,
          showOverflowTooltip: true,
          formatter: (row: RefundItem) => row.errorMsg || '-'
        },
        { prop: 'successTime', label: '退款时间', width: 170 },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 110,
          fixed: 'right',
          formatter: (row: RefundItem) =>
            h(ArtButtonTable, {
              type: 'action',
              icon: 'ri:eye-line',
              text: '详情',
              onClick: () => showDetail(row)
            })
        }
      ]
    }
  })

  const handleSearch = (params: Api.Pay.RefundSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开退款单详情：列表不展示渠道退款单号与失败原因全文，详情里看完整信息
  const showDetail = (row: RefundItem): void => {
    currentRefund.value = row
    detailVisible.value = true
  }
</script>
