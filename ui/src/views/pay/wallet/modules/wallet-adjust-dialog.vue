<!-- 钱包资金操作弹窗：充值 / 扣减 / 冻结 / 解冻，金额按「元」输入、按「分」提交 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="钱包资金操作"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      show-icon
      class="mb-5"
      title="操作立即生效并写入钱包流水"
    >
      <span class="text-xs">
        当前会员：{{ walletData?.userId || '-' }}，可用余额
        {{ formatPrice(walletData?.balance) }}，冻结金额
        {{ formatPrice(walletData?.frozenBalance) }}。 扣减 /
        冻结不能超过可用余额，解冻不能超过冻结金额。
      </span>
    </ElAlert>

    <ElForm ref="formRef" :model="form" :rules="rules" label-position="top">
      <ElFormItem label="操作类型" prop="direction">
        <ElRadioGroup v-model="form.direction">
          <ElRadioButton value="in">充值</ElRadioButton>
          <ElRadioButton value="out">扣减</ElRadioButton>
          <ElRadioButton value="freeze">冻结</ElRadioButton>
          <ElRadioButton value="unfreeze">解冻</ElRadioButton>
        </ElRadioGroup>
      </ElFormItem>

      <ElFormItem label="金额" prop="amountYuan">
        <ElInput v-model="form.amountYuan" placeholder="请输入金额（元，最多两位小数）">
          <template #append>元</template>
        </ElInput>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="3"
          maxlength="255"
          show-word-limit
          :placeholder="REMARK_PLACEHOLDER[form.direction]"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确认</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import {
    fetchAdjustPayWallet,
    fetchFreezePayWallet,
    fetchUnfreezePayWallet
  } from '@/api/pay/wallet'

  defineOptions({ name: 'WalletAdjustDialog' })

  type WalletItem = Api.Pay.WalletItem
  type Direction = 'in' | 'out' | 'freeze' | 'unfreeze'

  interface Props {
    visible: boolean
    walletData?: WalletItem | null
  }

  const props = withDefaults(defineProps<Props>(), { walletData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const REMARK_PLACEHOLDER: Record<Direction, string> = {
    in: '如：活动赠送 / 线下充值',
    out: '如：客诉补偿 / 线下退款',
    freeze: '如：风控冻结 / 争议订单待处理',
    unfreeze: '如：风控解除 / 争议处理完成'
  }

  const SUCCESS_TEXT: Record<Direction, string> = {
    in: '充值成功',
    out: '扣减成功',
    freeze: '冻结成功',
    unfreeze: '解冻成功'
  }

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    direction: 'in' as Direction,
    amountYuan: '',
    remark: ''
  })

  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`

  const validateAmount = (_rule: any, value: string, callback: (error?: Error) => void): void => {
    const amount = Number(value)
    if (!value || Number.isNaN(amount) || amount <= 0) {
      callback(new Error('请输入大于 0 的金额'))

      return
    }
    if (!/^\d+(\.\d{1,2})?$/.test(value)) {
      callback(new Error('金额最多保留两位小数'))

      return
    }

    const cents = Math.round(amount * 100)
    if (
      (form.direction === 'out' || form.direction === 'freeze') &&
      cents > (props.walletData?.balance ?? 0)
    ) {
      callback(new Error('金额不能超过当前可用余额'))

      return
    }
    if (form.direction === 'unfreeze' && cents > (props.walletData?.frozenBalance ?? 0)) {
      callback(new Error('解冻金额不能超过当前冻结金额'))

      return
    }

    callback()
  }

  const rules: FormRules = {
    amountYuan: [{ required: true, validator: validateAmount, trigger: 'blur' }]
  }

  // 切换操作类型后重新校验，避免沿用上一次的余额判断
  watch(
    () => form.direction,
    () => formRef.value?.clearValidate('amountYuan')
  )

  // 关闭后重置，避免下次打开残留上次的金额
  watch(dialogVisible, (visible) => {
    if (visible) return
    Object.assign(form, { direction: 'in', amountYuan: '', remark: '' })
    formRef.value?.clearValidate()
  })

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid || !props.walletData) return

    // 元 → 分；调账用正负号表达增减，冻结 / 解冻金额恒为正
    const amount = Math.round(Number(form.amountYuan) * 100)
    const payload = {
      userId: props.walletData.userId,
      amount,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (form.direction === 'in' || form.direction === 'out') {
        await fetchAdjustPayWallet({
          ...payload,
          amount: form.direction === 'out' ? -amount : amount
        })
      } else if (form.direction === 'freeze') {
        await fetchFreezePayWallet(payload)
      } else {
        await fetchUnfreezePayWallet(payload)
      }

      ElMessage.success(SUCCESS_TEXT[form.direction])
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
