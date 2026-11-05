<!-- 通知模板弹窗：内容支持 {{变量}} 占位 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑通知模板' : '新增通知模板'"
    width="680px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-position="top" class="notice-form">
      <div class="form-row-2">
        <ElFormItem label="模板编码" prop="code">
          <ElInput
            v-model="form.code"
            :disabled="isEdit"
            placeholder="业务调用凭据，如 order.paid"
            maxlength="64"
          />
        </ElFormItem>

        <ElFormItem label="模板名称" prop="name">
          <ElInput v-model="form.name" placeholder="如：订单支付成功" maxlength="64" />
        </ElFormItem>

        <ElFormItem label="渠道类型" prop="type">
          <ElSelect v-model="form.type" class="w-full">
            <ElOption
              v-for="item in types"
              :key="item.type"
              :label="item.name"
              :value="item.type"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="状态" prop="status">
          <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
          <span class="ml-2 text-xs text-g-600">{{ form.status === 1 ? '启用' : '停用' }}</span>
        </ElFormItem>
      </div>

      <ElFormItem label="标题" prop="title">
        <ElInput v-model="form.title" placeholder="邮件主题 / 机器人消息标题" maxlength="128" />
      </ElFormItem>

      <ElFormItem label="内容" prop="content">
        <ElInput
          v-model="form.content"
          type="textarea"
          :autosize="{ minRows: 6, maxRows: 14 }"
          maxlength="20000"
          show-word-limit
          placeholder="支持 {{变量}} 占位，如：订单 {{orderNo}} 已支付 {{amount}} 元"
        />
        <div class="mt-1 text-xs text-g-600">
          发送时传入同名变量即可替换；未传的变量保持原样，便于排查漏传
        </div>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput v-model="form.remark" placeholder="可选" maxlength="255" />
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
  import { fetchCreateNoticeTemplate, fetchUpdateNoticeTemplate } from '@/api/infra/notice'

  defineOptions({ name: 'NoticeTemplateDialog' })

  type TemplateItem = Api.Infra.NoticeTemplateItem
  type TypeSchema = Api.Infra.NoticeTypeSchema

  interface Props {
    visible: boolean
    templateData?: TemplateItem | null
    types: TypeSchema[]
  }

  const props = withDefaults(defineProps<Props>(), { templateData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => !!props.templateData)

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    code: '',
    name: '',
    type: '',
    title: '',
    content: '',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    code: [
      { required: true, message: '请输入模板编码', trigger: 'blur' },
      { max: 64, message: '模板编码不能超过 64 个字符', trigger: 'blur' }
    ],
    name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
    type: [{ required: true, message: '请选择渠道类型', trigger: 'change' }],
    content: [{ required: true, message: '请填写模板内容', trigger: 'blur' }]
  }

  watch(
    () => [props.visible, props.templateData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.templateData
      Object.assign(form, {
        code: data?.code ?? '',
        name: data?.name ?? '',
        type: data?.type ?? props.types[0]?.type ?? '',
        title: data?.title ?? '',
        content: data?.content ?? '',
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      const payload: Api.Infra.NoticeTemplateSaveParams = {
        code: form.code.trim(),
        name: form.name.trim(),
        type: form.type,
        title: form.title.trim(),
        content: form.content,
        status: form.status,
        remark: form.remark.trim()
      }

      if (props.templateData) {
        await fetchUpdateNoticeTemplate(props.templateData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateNoticeTemplate(payload)
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
