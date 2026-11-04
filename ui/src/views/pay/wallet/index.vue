<!-- 会员钱包：余额总览 + 后台调账 + 流水查询 -->
<template>
  <div class="pay-wallet-page art-full-height">
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
            <ElTag type="info" effect="plain"> 余额单位：元；会员首次访问钱包时自动建档 </ElTag>
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

    <WalletAdjustDialog
      v-model:visible="adjustVisible"
      :wallet-data="currentWallet"
      @success="refreshUpdate"
    />
    <WalletRecordsDrawer v-model:visible="recordsVisible" :wallet-data="currentWallet" />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetPayWalletList } from '@/api/pay/wallet'
  import { ElCard, ElSpace, ElTag } from 'element-plus'
  import WalletAdjustDialog from './modules/wallet-adjust-dialog.vue'
  import WalletRecordsDrawer from './modules/wallet-records-drawer.vue'

  defineOptions({ name: 'PayWallet' })

  type WalletItem = Api.Pay.WalletItem

  const searchForm = ref<Api.Pay.WalletSearchParams>({
    userId: undefined
  })

  const searchItems = [
    {
      label: '会员 ID',
      key: 'userId',
      type: 'input',
      placeholder: '请输入会员 ID',
      clearable: true
    }
  ]

  const adjustVisible = ref(false)
  const recordsVisible = ref(false)
  const currentWallet = ref<WalletItem | null>(null)

  // 金额（分）转「元」
  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`

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
      apiFn: fetchGetPayWalletList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'userId', label: '会员 ID', minWidth: 190, showOverflowTooltip: true },
        {
          prop: 'balance',
          label: '可用余额',
          width: 130,
          formatter: (row: WalletItem) =>
            h('span', { class: 'font-medium text-[15px]' }, formatPrice(row.balance))
        },
        {
          prop: 'frozenBalance',
          label: '冻结金额',
          width: 130,
          formatter: (row: WalletItem) =>
            h(
              'span',
              { class: row.frozenBalance > 0 ? 'text-warning' : '' },
              formatPrice(row.frozenBalance)
            )
        },
        {
          prop: 'totalRecharge',
          label: '累计充值',
          width: 120,
          formatter: (row: WalletItem) => formatPrice(row.totalRecharge)
        },
        {
          prop: 'totalExpense',
          label: '累计消费',
          width: 130,
          formatter: (row: WalletItem) => formatPrice(row.totalExpense)
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: WalletItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '正常' : '冻结'
            )
        },
        { prop: 'updatedAt', label: '最近变动', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 200,
          fixed: 'right',
          formatter: (row: WalletItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'edit',
                text: '资金操作',
                onClick: () => showAdjust(row)
              }),
              h(ArtButtonTable, {
                type: 'action',
                icon: 'ri:file-list-3-line',
                text: '流水',
                onClick: () => showRecords(row)
              })
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Pay.WalletSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showAdjust = (row: WalletItem): void => {
    currentWallet.value = row
    adjustVisible.value = true
  }

  const showRecords = (row: WalletItem): void => {
    currentWallet.value = row
    recordsVisible.value = true
  }
</script>
