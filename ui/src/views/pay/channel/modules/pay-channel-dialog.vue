<!-- 支付渠道新增 / 编辑弹窗：按后端下发的字段定义动态渲染表单 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑支付渠道' : '新增支付渠道'"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="130px">
      <ElFormItem label="渠道编码" prop="code">
        <ElSelect
          v-model="form.code"
          :disabled="isEdit"
          placeholder="请选择支付渠道"
          class="w-full"
          @change="handleCodeChange"
        >
          <ElOption
            v-for="item in codeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <div class="mt-1 text-xs text-g-600">
          一个渠道编码对应一种支付方式；同一商户号要同时支持扫码与 H5，请建两条记录
        </div>
      </ElFormItem>

      <ElFormItem v-if="notifyUrl" label="异步通知地址">
        <div class="flex w-full items-center gap-2">
          <ElInput :model-value="notifyUrl" readonly />
          <ElButton v-ripple @click="handleCopyNotifyUrl">复制</ElButton>
        </div>
        <div class="mt-1 text-xs text-g-600">
          由系统按 APP_URL 自动生成，无需填写；请确保该域名公网可访问且为 HTTPS
        </div>
      </ElFormItem>

      <ElFormItem label="渠道名称" prop="name">
        <ElInput v-model="form.name" placeholder="如：微信小程序支付" maxlength="64" />
      </ElFormItem>

      <template v-for="field in fields" :key="field.key">
        <ElFormItem :label="field.label">
          <ElSwitch
            v-if="field.type === 'switch'"
            v-model="configValues[field.key]"
            :active-value="true"
            :inactive-value="false"
          />
          <ElInput
            v-else-if="field.type === 'textarea'"
            v-model="configValues[field.key]"
            type="textarea"
            :rows="field.rows || 4"
            :placeholder="fieldPlaceholder(field)"
          />
          <ElInput
            v-else
            v-model="configValues[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :show-password="field.type === 'password'"
            :placeholder="fieldPlaceholder(field)"
            :maxlength="255"
          />
          <div v-if="field.desc" class="mt-1 text-xs text-g-600">{{ field.desc }}</div>
        </ElFormItem>
      </template>

      <ElFormItem label="状态" prop="status">
        <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
        <span class="ml-2 text-xs text-g-600">停用后用户将无法选择该支付方式</span>
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        <span class="ml-2 text-xs text-g-600">数字越小越靠前</span>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput v-model="form.remark" type="textarea" :rows="2" maxlength="255" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElDialog, ElMessage } from 'element-plus'
  import {
    fetchCreatePayChannel,
    fetchGetPayChannelNotifyUrl,
    fetchGetPayChannelSchema,
    fetchUpdatePayChannel
  } from '@/api/pay/channel'

  defineOptions({ name: 'PayChannelDialog' })

  type ChannelItem = Api.Pay.ChannelItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    channelData?: ChannelItem | null
  }

  const props = withDefaults(defineProps<Props>(), { channelData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  // 渠道编码清单：字段定义由后端下发，编码枚举前端固定，避免空 code 调 schema
  const codeOptions = [
    { label: '微信扫码（Native）', value: 'wx_native' },
    { label: '微信 H5', value: 'wx_h5' },
    { label: '微信 JSAPI（公众号）', value: 'wx_jsapi' },
    { label: '微信小程序', value: 'wx_mini' },
    { label: '微信 APP', value: 'wx_app' },
    { label: '支付宝电脑网站', value: 'ali_pc' },
    { label: '支付宝手机网站', value: 'ali_wap' },
    { label: '支付宝 APP', value: 'ali_app' },
    { label: '支付宝当面付扫码', value: 'ali_qr' }
  ]

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const fields = ref<Api.Pay.ChannelField[]>([])
  const notifyUrl = ref('')

  const form = reactive({
    name: '',
    code: '',
    status: 1,
    sort: 0,
    remark: ''
  })

  // 渠道 config 的动态字段值：string 或 boolean（switch）
  const configValues = reactive<Record<string, any>>({})

  const rules: FormRules = {
    code: [{ required: true, message: '请选择支付渠道', trigger: 'change' }],
    name: [
      { required: true, message: '请输入渠道名称', trigger: 'blur' },
      { max: 64, message: '渠道名称不能超过 64 个字符', trigger: 'blur' }
    ]
  }

  // 动态字段占位提示：敏感字段在编辑态明确「留空不修改」
  const fieldPlaceholder = (field: Api.Pay.ChannelField): string => {
    if (field.secret && isEdit.value) {
      return '留空表示不修改'
    }
    return field.required ? `请输入${field.label}（必填）` : `请输入${field.label}（选填）`
  }

  // 清空并重建动态字段默认值
  const resetConfigValues = (): void => {
    Object.keys(configValues).forEach((key) => delete configValues[key])
    fields.value.forEach((field) => {
      configValues[field.key] = field.type === 'switch' ? false : ''
    })
  }

  // 按渠道编码拉取字段定义与通知地址
  const loadSchema = async (code: string): Promise<void> => {
    if (!code) {
      fields.value = []
      notifyUrl.value = ''
      return
    }

    try {
      const schema = await fetchGetPayChannelSchema(code)
      fields.value = schema.fields ?? []
      resetConfigValues()
    } catch {
      fields.value = []
    }

    try {
      const res = await fetchGetPayChannelNotifyUrl(code)
      notifyUrl.value = res?.payNotifyUrl ?? ''
    } catch {
      notifyUrl.value = ''
    }
  }

  const handleCodeChange = (code: string): void => {
    loadSchema(code)
  }

  // 打开弹窗时回填：敏感字段一律留空（脱敏值不能回写），其余按 JSON 还原
  watch(
    () => [props.visible, props.channelData] as const,
    async ([visible]) => {
      if (!visible) return

      const data = props.channelData
      Object.assign(form, {
        name: data?.name ?? '',
        code: data?.code ?? '',
        status: data?.status ?? 1,
        sort: data?.sort ?? 0,
        remark: data?.remark ?? ''
      })

      await loadSchema(form.code)

      if (!data?.config) {
        nextTick(() => formRef.value?.clearValidate())
        return
      }

      let conf: Record<string, any> = {}
      try {
        conf = JSON.parse(data.config) ?? {}
      } catch {
        conf = {}
      }

      fields.value.forEach((field) => {
        if (field.type === 'switch') {
          configValues[field.key] = conf[field.key] === true
          return
        }
        // 敏感字段不回填，提交时留空由后端沿用原值
        configValues[field.key] = field.secret ? '' : (conf[field.key] ?? '')
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 组装 config JSON：按字段定义逐个取值，未定义的键不提交，避免脏数据
  const buildConfig = (): string => {
    const conf: Record<string, any> = {}
    fields.value.forEach((field) => {
      const value = configValues[field.key]
      conf[field.key] = field.type === 'switch' ? value === true : String(value ?? '').trim()
    })
    return JSON.stringify(conf)
  }

  // 提交前校验必填项：编辑态允许敏感字段留空（沿用原值）
  const validateRequiredFields = (): boolean => {
    for (const field of fields.value) {
      if (!field.required) continue
      if (field.secret && isEdit.value) continue
      const value = configValues[field.key]
      if (field.type === 'switch') continue
      if (String(value ?? '').trim() === '') {
        ElMessage.warning(`${field.label}不能为空`)
        return false
      }
    }
    return true
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return
    if (!validateRequiredFields()) return

    const payload: Api.Pay.ChannelSaveParams = {
      name: form.name.trim(),
      code: form.code,
      config: buildConfig(),
      status: form.status,
      sort: form.sort,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.channelData) {
        await fetchUpdatePayChannel(props.channelData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreatePayChannel(payload)
        ElMessage.success('新增成功')
      }

      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  // 复制异步通知地址，供用户粘贴到微信 / 支付宝商户平台核对域名
  const handleCopyNotifyUrl = async (): Promise<void> => {
    if (!notifyUrl.value) return
    try {
      await navigator.clipboard.writeText(notifyUrl.value)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动选中复制')
    }
  }
</script>
