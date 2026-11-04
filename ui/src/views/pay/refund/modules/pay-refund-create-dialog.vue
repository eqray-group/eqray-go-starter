<!-- 发起退款弹窗：按支付单号定位订单，填写金额与原因后直接调渠道 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="发起退款"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="warning"
      :closable="false"
      show-icon
      class="refund-alert"
      title="提交后立即调用渠道退款接口，无人工复核"
    >
      <span class="text-xs">
        请先核对支付单号与退款金额；支持部分退款、可多次提交，累计不超过订单金额。
      </span>
    </ElAlert>

    <ElForm ref="formRef" :model="form" :rules="rules" label-position="top" class="refund-form">
      <ElFormItem label="支付单号" prop="merchantOrderNo">
        <div class="flex w-full items-center gap-2">
          <ElInput
            v-model="form.merchantOrderNo"
            placeholder="请输入已支付的支付单号"
            maxlength="64"
            @blur="handleQueryOrder"
          />
          <ElButton :loading="querying" @click="handleQueryOrder">查询</ElButton>
        </div>
      </ElFormItem>

      <transition name="el-fade-in">
        <div v-if="order" class="order-panel">
          <div class="order-row">
            <span class="order-label">业务单号</span>
            <span class="order-value">{{ order.bizType }} / {{ order.bizOrderNo }}</span>
          </div>
          <div class="order-row">
            <span class="order-label">订单金额</span>
            <span class="order-value is-strong">{{ formatPrice(order.price) }}</span>
          </div>
          <div class="order-row">
            <span class="order-label">订单状态</span>
            <ElTag :type="order.status === 20 ? 'success' : 'warning'" size="small">
              {{ STATUS_LABEL[order.status] || '未知' }}
            </ElTag>
          </div>
        </div>
      </transition>

      <ElFormItem label="退款金额" prop="refundYuan">
        <ElInput v-model="form.refundYuan" placeholder="请输入退款金额（元，最多两位小数）">
          <template #append>元</template>
        </ElInput>
        <div class="mt-1 text-xs text-g-600">支持部分退款，累计退款不能超过订单金额</div>
      </ElFormItem>

      <ElFormItem label="退款原因" prop="reason">
        <ElInput
          v-model="form.reason"
          type="textarea"
          :rows="3"
          maxlength="255"
          show-word-limit
          placeholder="如：用户申请退款"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交退款</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElDialog, ElMessage } from 'element-plus'
  import { fetchGetPayOrderList } from '@/api/pay/order'
  import { fetchCreatePayRefund } from '@/api/pay/refund'

  defineOptions({ name: 'PayRefundCreateDialog' })

  type OrderItem = Api.Pay.OrderItem

  interface Props {
    visible: boolean
  }

  const props = defineProps<Props>()

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  // 订单状态文案，用于提示「未支付的单不能退款」
  const STATUS_LABEL: Record<number, string> = {
    0: '待支付',
    10: '支付中',
    20: '已支付',
    30: '已关闭'
  }

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const querying = ref(false)
  const order = ref<OrderItem | null>(null)

  const form = reactive({
    merchantOrderNo: '',
    refundYuan: '',
    reason: ''
  })

  // 金额（分）转「元」
  const formatPrice = (price: number): string => `${((price || 0) / 100).toFixed(2)} 元`

  // 退款金额校验：必须是大于 0 且最多两位小数的数字
  const validateAmount = (_rule: any, value: string, callback: (error?: Error) => void): void => {
    const amount = Number(value)
    if (!value || Number.isNaN(amount) || amount <= 0) {
      callback(new Error('请输入大于 0 的退款金额'))
      return
    }
    if (!/^\d+(\.\d{1,2})?$/.test(value)) {
      callback(new Error('退款金额最多保留两位小数'))
      return
    }
    callback()
  }

  const rules: FormRules = {
    merchantOrderNo: [{ required: true, message: '请输入支付单号', trigger: 'blur' }],
    refundYuan: [{ required: true, validator: validateAmount, trigger: 'blur' }],
    reason: [{ required: true, message: '请输入退款原因', trigger: 'blur' }]
  }

  // 关闭时重置，避免下次打开残留上次的订单与金额
  watch(dialogVisible, (visible) => {
    if (visible) return
    order.value = null
    Object.assign(form, { merchantOrderNo: '', refundYuan: '', reason: '' })
  })

  // 按支付单号查询订单：确认存在且已支付，避免用户填错单号
  const handleQueryOrder = async (): Promise<void> => {
    const merchantOrderNo = form.merchantOrderNo.trim()
    if (!merchantOrderNo) return

    querying.value = true
    try {
      const res = await fetchGetPayOrderList({ merchantOrderNo, pageNo: 1, pageSize: 1 })
      const record = res?.list?.[0] ?? null
      order.value = record
      if (!record) {
        ElMessage.warning('未找到该支付订单')
      } else if (record.status !== 20) {
        ElMessage.warning('该订单未支付成功，不能退款')
      }
    } catch {
      order.value = null
      // 错误提示已由请求层统一弹出
    } finally {
      querying.value = false
    }
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    if (!order.value) {
      await handleQueryOrder()
    }
    if (!order.value) return
    if (order.value.status !== 20) {
      ElMessage.warning('该订单未支付成功，不能退款')
      return
    }

    // 元 → 分，避免浮点误差
    const refundPrice = Math.round(Number(form.refundYuan) * 100)
    if (refundPrice > order.value.price) {
      ElMessage.warning('退款金额不能超过订单金额')
      return
    }

    submitting.value = true
    try {
      await fetchCreatePayRefund({
        orderId: order.value.id,
        refundPrice,
        reason: form.reason.trim()
      })
      ElMessage.success('退款已提交')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .refund-alert {
    margin-bottom: 22px;
    border-radius: 8px;
  }

  .refund-form {
    :deep(.el-form-item) {
      margin-bottom: 22px;
    }

    :deep(.el-form-item__label) {
      padding-bottom: 6px;
      font-weight: 500;
      color: var(--el-text-color-regular);
    }
  }

  // 订单信息块：替代原来挤在表单行里的三行纯文本
  .order-panel {
    margin-bottom: 22px;
    padding: 12px 16px;
    border-radius: 8px;
    background: var(--el-fill-color-lighter);
  }

  .order-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;

    & + & {
      margin-top: 10px;
    }
  }

  .order-label {
    flex-shrink: 0;
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .order-value {
    overflow: hidden;
    font-size: 13px;
    color: var(--el-text-color-regular);
    text-align: right;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.is-strong {
      font-weight: 600;
      color: var(--el-color-danger);
    }
  }
</style>
