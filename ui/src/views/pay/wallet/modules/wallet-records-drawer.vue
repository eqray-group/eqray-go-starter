<!-- 钱包流水抽屉：充值 / 消费 / 退款 / 调整的完整记录，用于对账 -->
<template>
  <ElDrawer v-model="drawerVisible" title="钱包流水" size="760px" @open="loadRecords">
    <ElDescriptions :column="4" border class="mb-4">
      <ElDescriptionsItem label="会员 ID">{{ walletData?.userId || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="可用余额">
        {{ formatPrice(walletData?.balance) }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="冻结金额">
        {{ formatPrice(walletData?.frozenBalance) }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="累计消费">
        {{ formatPrice(walletData?.totalExpense) }}
      </ElDescriptionsItem>
    </ElDescriptions>

    <ElTable v-loading="loading" :data="records" border stripe>
      <ElTableColumn prop="createdAt" label="时间" width="170" />
      <ElTableColumn label="类型" width="90">
        <template #default="{ row }">
          <ElTag :type="TYPE_TAG[row.type] || 'info'" size="small">
            {{ TYPE_TEXT[row.type] || '未知' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="金额" width="120">
        <template #default="{ row }">
          <span :class="row.amount >= 0 ? 'text-success' : 'text-danger'">
            {{ row.amount >= 0 ? '+' : '' }}{{ formatPrice(row.amount) }}
          </span>
        </template>
      </ElTableColumn>
      <ElTableColumn label="变动后余额" width="130">
        <template #default="{ row }">{{ formatPrice(row.balanceAfter) }}</template>
      </ElTableColumn>
      <ElTableColumn prop="bizOrderNo" label="关联单号" min-width="170" show-overflow-tooltip />
      <ElTableColumn prop="remark" label="备注" min-width="160" show-overflow-tooltip />
    </ElTable>

    <div class="mt-4 flex justify-end">
      <ElPagination
        v-model:current-page="pageNo"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadRecords"
        @size-change="loadRecords"
      />
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import {
    ElDescriptions,
    ElDescriptionsItem,
    ElDrawer,
    ElPagination,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import { fetchGetPayWalletRecords } from '@/api/pay/wallet'

  defineOptions({ name: 'WalletRecordsDrawer' })

  type WalletItem = Api.Pay.WalletItem
  type WalletRecordItem = Api.Pay.WalletRecordItem

  interface Props {
    visible: boolean
    walletData?: WalletItem | null
  }

  const props = withDefaults(defineProps<Props>(), { walletData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  // 流水类型文案与配色
  const TYPE_TEXT: Record<number, string> = {
    1: '充值',
    2: '消费',
    3: '退款',
    4: '调整',
    5: '冻结',
    6: '解冻'
  }
  const TYPE_TAG: Record<number, 'success' | 'danger' | 'warning' | 'info'> = {
    1: 'success',
    2: 'danger',
    3: 'warning',
    4: 'info',
    5: 'warning',
    6: 'success'
  }

  const loading = ref(false)
  const records = ref<WalletRecordItem[]>([])
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = ref(20)

  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`

  const loadRecords = async (): Promise<void> => {
    if (!props.walletData?.userId) return

    loading.value = true
    try {
      const res = await fetchGetPayWalletRecords({
        userId: props.walletData.userId,
        pageNo: pageNo.value,
        pageSize: pageSize.value
      })
      records.value = res?.list ?? []
      total.value = res?.total ?? 0
    } catch {
      records.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }
</script>
