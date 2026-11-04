<!-- 退款单详情弹窗：只读展示退款单全量信息，失败单重点看失败原因 -->
<template>
  <ElDialog v-model="dialogVisible" title="退款单详情" width="720px" align-center>
    <ElDescriptions :column="2" border>
      <ElDescriptionsItem label="退款单号" :span="2">
        {{ refund?.merchantRefundNo || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="支付单号" :span="2">
        {{ refund?.merchantOrderNo || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="支付渠道">{{ refund?.channelCode || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="退款金额">{{
        formatPrice(refund?.refundPrice)
      }}</ElDescriptionsItem>
      <ElDescriptionsItem label="退款状态">
        <ElTag :type="statusTag.type">{{ statusTag.label }}</ElTag>
      </ElDescriptionsItem>
      <ElDescriptionsItem label="渠道退款单号">
        {{ refund?.channelRefundNo || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="退款原因" :span="2">{{
        refund?.reason || '-'
      }}</ElDescriptionsItem>
      <ElDescriptionsItem v-if="refund?.errorMsg" label="失败原因" :span="2">
        <span class="text-danger">{{ refund?.errorMsg }}</span>
      </ElDescriptionsItem>
      <ElDescriptionsItem label="退款时间">{{ refund?.successTime || '-' }}</ElDescriptionsItem>
      <ElDescriptionsItem label="创建时间">{{ refund?.createdAt || '-' }}</ElDescriptionsItem>
    </ElDescriptions>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElDescriptions, ElDescriptionsItem, ElDialog, ElTag } from 'element-plus'

  defineOptions({ name: 'PayRefundDetailDialog' })

  type RefundItem = Api.Pay.RefundItem

  interface Props {
    visible: boolean
    refundData?: RefundItem | null
  }

  const props = withDefaults(defineProps<Props>(), { refundData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const refund = computed(() => props.refundData)

  // 退款状态文案，与列表页保持一致
  const STATUS_MAP: Record<
    number,
    { label: string; type: 'info' | 'warning' | 'success' | 'danger' }
  > = {
    0: { label: '待退款', type: 'info' },
    10: { label: '退款中', type: 'warning' },
    20: { label: '已退款', type: 'success' },
    30: { label: '退款失败', type: 'danger' }
  }

  const statusTag = computed(
    () => STATUS_MAP[refund.value?.status ?? -1] ?? { label: '未知', type: 'info' as const }
  )

  // 金额（分）转「元」
  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`
</script>
