<!-- 通知渠道弹窗：配置字段由后端类型定义动态渲染 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑通知渠道' : '新增通知渠道'"
    width="640px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-position="top" class="notice-form">
      <div class="form-row-2">
        <ElFormItem label="渠道类型" prop="type">
          <ElSelect
            v-model="form.type"
            :disabled="isEdit"
            class="w-full"
            @change="handleTypeChange"
          >
            <ElOption
              v-for="item in types"
              :key="item.type"
              :label="item.name"
              :value="item.type"
            />
          </ElSelect>
          <div v-if="currentType" class="mt-1 text-xs text-g-600">{{ currentType.desc }}</div>
        </ElFormItem>

        <ElFormItem label="渠道名称" prop="name">
          <ElInput v-model="form.name" placeholder="如：运营告警群 / 系统邮件" maxlength="64" />
        </ElFormItem>
      </div>

      <ElDivider class="!my-2" />

      <template v-for="field in currentType?.fields ?? []" :key="field.key">
        <ElFormItem :label="field.label" :required="field.required">
          <ElSelect
            v-if="field.type === 'select'"
            v-model="configValues[field.key]"
            class="w-full"
            :placeholder="field.placeholder || '请选择'"
          >
            <ElOption
              v-for="option in field.options ?? []"
              :key="option.value"
              :label="option.label"
              :value="option.value"
            />
          </ElSelect>

          <ElInput
            v-else
            v-model="configValues[field.key]"
            :type="field.type === 'password' ? 'password' : 'text'"
            :show-password="field.type === 'password'"
            :placeholder="
              field.placeholder || (field.type === 'password' && isEdit ? '留空表示不修改' : '')
            "
            :maxlength="field.type === 'text' ? 255 : undefined"
          />

          <div v-if="field.desc" class="mt-1 text-xs text-g-600">{{ field.desc }}</div>
        </ElFormItem>
      </template>

      <div class="form-row-2">
        <ElFormItem label="状态" prop="status">
          <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
          <span class="ml-2 text-xs text-g-600">{{ form.status === 1 ? '启用' : '停用' }}</span>
        </ElFormItem>

        <ElFormItem label="备注" prop="remark">
          <ElInput v-model="form.remark" placeholder="可选" maxlength="255" />
        </ElFormItem>
      </div>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElDialog, ElDivider, ElMessage } from 'element-plus'
  import { fetchCreateNoticeChannel, fetchUpdateNoticeChannel } from '@/api/infra/notice'

  defineOptions({ name: 'NoticeChannelDialog' })

  type ChannelItem = Api.Infra.NoticeChannelItem
  type TypeSchema = Api.Infra.NoticeTypeSchema

  interface Props {
    visible: boolean
    channelData?: ChannelItem | null
    types: TypeSchema[]
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

  const isEdit = computed(() => !!props.channelData)

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const configValues = ref<Record<string, string>>({})

  const form = reactive({
    name: '',
    type: '',
    status: 1,
    remark: ''
  })

  const currentType = computed(() => props.types.find((item) => item.type === form.type) || null)

  const rules: FormRules = {
    type: [{ required: true, message: '请选择渠道类型', trigger: 'change' }],
    name: [
      { required: true, message: '请输入渠道名称', trigger: 'blur' },
      { max: 64, message: '渠道名称不能超过 64 个字符', trigger: 'blur' }
    ]
  }

  // 把库里的 JSON 文本解析成表单值
  const parseConfig = (raw?: string): Record<string, string> => {
    if (!raw) return {}

    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>
      const result: Record<string, string> = {}
      Object.keys(parsed).forEach((key) => {
        result[key] = parsed[key] == null ? '' : String(parsed[key])
      })

      return result
    } catch {
      return {}
    }
  }

  // 切换类型时重置配置并给加密方式一个默认值
  const handleTypeChange = (type: string): void => {
    configValues.value = {}
    const schema = props.types.find((item) => item.type === type)
    if (!schema) return

    schema.fields.forEach((field) => {
      if (field.type === 'select' && field.options?.length) {
        configValues.value[field.key] = field.options[0].value
      }
      if (field.key === 'port' && type === 'email') configValues.value.port = '465'
    })
  }

  watch(
    () => [props.visible, props.channelData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.channelData
      Object.assign(form, {
        name: data?.name ?? '',
        type: data?.type ?? '',
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      configValues.value = data ? parseConfig(data.config) : {}

      if (!data && !form.type && props.types.length) {
        form.type = props.types[0].type
        handleTypeChange(form.type)
      }

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const schema = currentType.value
    if (!schema) return

    // 前端先挡一次必填，避免只靠后端提示
    const missing = schema.fields.filter(
      (field) => field.required && !String(configValues.value[field.key] ?? '').trim()
    )
    if (missing.length > 0) {
      ElMessage.warning(`请填写：${missing.map((item) => item.label).join('、')}`)
      return
    }

    submitting.value = true
    try {
      const payload: Api.Infra.NoticeChannelSaveParams = {
        name: form.name.trim(),
        type: form.type,
        config: { ...configValues.value },
        status: form.status,
        remark: form.remark.trim()
      }

      if (props.channelData) {
        await fetchUpdateNoticeChannel(props.channelData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateNoticeChannel(payload)
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
</script>

<style lang="scss" scoped>
  .notice-form {
    :deep(.el-form-item) {
      margin-bottom: 16px;
    }

    :deep(.el-form-item__label) {
      padding-bottom: 4px;
      font-weight: 500;
    }
  }

  .form-row-2 {
    display: grid;
    gap: 0 16px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
</style>
