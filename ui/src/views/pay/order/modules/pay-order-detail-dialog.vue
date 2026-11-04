<!-- 支付订单详情弹窗：订单信息 + 该单的渠道回调报文（排查未到账时不用再跳页） -->
<template>
  <ElDialog v-model="dialogVisible" title="支付订单详情" width="860px" align-center>
    <ElTabs v-model="activeTab">
      <ElTabPane label="订单信息" name="order">
        <ElDescriptions :column="2" border>
          <ElDescriptionsItem label="支付单号" :span="2">
            {{ order?.merchantOrderNo || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="业务类型">{{ order?.bizType || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="业务单号">{{ order?.bizOrderNo || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="商品标题" :span="2">
            {{ order?.subject || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="支付渠道">{{ order?.channelCode || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="金额">{{ formatPrice(order?.price) }}</ElDescriptionsItem>
          <ElDescriptionsItem label="订单状态">
            <ElTag :type="statusTag.type">{{ statusTag.label }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem label="退款进度">{{ refundLabel }}</ElDescriptionsItem>
          <ElDescriptionsItem label="渠道单号">{{
            order?.channelOrderNo || '-'
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="付款人标识">{{
            order?.channelUser || '-'
          }}</ElDescriptionsItem>
          <ElDescriptionsItem label="下单会员">{{ order?.userId || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="下单 IP">{{ order?.clientIp || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="支付超时">{{ order?.expireTime || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="支付时间">{{ order?.successTime || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="创建时间">{{ order?.createdAt || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="更新时间">{{ order?.updatedAt || '-' }}</ElDescriptionsItem>
          <ElDescriptionsItem label="异步通知地址" :span="2">
            <span class="break-all">{{ order?.notifyUrl || '-' }}</span>
          </ElDescriptionsItem>
        </ElDescriptions>
      </ElTabPane>

      <ElTabPane :label="`回调报文（${logs.length}）`" name="notify">
        <ElEmpty v-if="!logs.length" description="该订单暂无回调记录" />
        <div v-for="item in logs" :key="item.id" class="mb-3">
          <div class="mb-1 flex items-center gap-2 text-xs text-g-600">
            <ElTag :type="item.verifyStatus === 1 ? 'success' : 'danger'" size="small">
              验签{{ item.verifyStatus === 1 ? '通过' : '失败' }}
            </ElTag>
            <ElTag :type="item.handleStatus === 1 ? 'success' : 'danger'" size="small">
              {{ item.handleStatus === 1 ? '处理成功' : '处理失败' }}
            </ElTag>
            <span>{{ item.createdAt }}</span>
            <span v-if="item.errorMsg" class="text-danger">{{ item.errorMsg }}</span>
          </div>
          <pre class="notify-body">{{ prettyBody(item.decryptedBody || item.requestBody) }}</pre>
        </div>
      </ElTabPane>
    </ElTabs>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    ElDescriptions,
    ElDescriptionsItem,
    ElDialog,
    ElEmpty,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import { fetchGetPayNotifyLogList } from '@/api/pay/notify-log'

  defineOptions({ name: 'PayOrderDetailDialog' })

  type OrderItem = Api.Pay.OrderItem
  type NotifyLogItem = Api.Pay.NotifyLogItem

  interface Props {
    visible: boolean
    orderData?: OrderItem | null
  }

  const props = withDefaults(defineProps<Props>(), { orderData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const order = computed(() => props.orderData)
  const activeTab = ref('order')
  const logs = ref<NotifyLogItem[]>([])

  // 订单状态文案与配色，与列表页保持一致
  const STATUS_MAP: Record<
    number,
    { label: string; type: 'info' | 'warning' | 'success' | 'danger' }
  > = {
    0: { label: '待支付', type: 'info' },
    10: { label: '支付中', type: 'warning' },
    20: { label: '已支付', type: 'success' },
    30: { label: '已关闭', type: 'danger' }
  }

  const statusTag = computed(
    () => STATUS_MAP[order.value?.status ?? -1] ?? { label: '未知', type: 'info' as const }
  )

  // 退款进度文案
  const REFUND_MAP: Record<number, string> = { 0: '未退款', 10: '部分退款', 20: '全额退款' }
  const refundLabel = computed(() => REFUND_MAP[order.value?.refundStatus ?? 0] ?? '-')

  // 金额（分）转「元」
  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`

  // 报文尽量格式化展示，非 JSON（支付宝表单）原样输出
  const prettyBody = (body?: string): string => {
    if (!body) return '（空）'
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }

  // 打开时按支付单号拉取该单的回调记录，便于直接确认"回调是否到达、是否验签通过"
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      activeTab.value = 'order'
      logs.value = []
      const merchantOrderNo = props.orderData?.merchantOrderNo
      if (!merchantOrderNo) return

      try {
        const res = await fetchGetPayNotifyLogList({ merchantOrderNo, pageNo: 1, pageSize: 20 })
        logs.value = res?.list ?? []
      } catch {
        // 日志拉取失败不影响订单信息展示，错误提示已由请求层统一弹出
      }
    }
  )
</script>

<style lang="scss" scoped>
  .notify-body {
    max-height: 260px;
    overflow: auto;
    margin: 0;
    padding: 10px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
