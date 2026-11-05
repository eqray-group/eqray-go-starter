<!-- 定时任务新增 / 编辑弹窗：处理器只能从代码注册的清单里选 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑定时任务' : '新增定时任务'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      type="info"
      :closable="false"
      show-icon
      class="job-alert"
      title="任务逻辑由代码注册，这里只配置频率与启停"
    >
      <span class="text-xs">修改执行频率、暂停任务无需发版；改逻辑仍需改代码里的处理器实现。</span>
    </ElAlert>

    <ElForm ref="formRef" :model="form" :rules="rules" label-position="top" class="job-form">
      <ElFormItem label="任务处理器" prop="handler">
        <ElSelect
          v-model="form.handler"
          :disabled="isEdit"
          placeholder="请选择处理器"
          class="w-full"
          @change="handleHandlerChange"
        >
          <ElOption
            v-for="item in handlers"
            :key="item.key"
            :label="`${item.name}（${item.key}）`"
            :value="item.key"
          />
        </ElSelect>
        <div v-if="currentHandler" class="mt-1 text-xs text-g-600">
          {{ currentHandler.description }}
        </div>
      </ElFormItem>

      <ElFormItem label="任务名称" prop="name">
        <ElInput v-model="form.name" placeholder="默认取处理器名称，可修改" maxlength="64" />
      </ElFormItem>

      <ElFormItem label="执行频率" prop="cron">
        <div class="flex w-full items-center gap-2">
          <ElSelect
            v-model="cronPreset"
            placeholder="常用频率"
            style="width: 160px"
            @change="handlePreset"
          >
            <ElOption
              v-for="item in CRON_PRESETS"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
          <ElInput v-model="form.cron" placeholder="cron 表达式，如 */5 * * * *" maxlength="64" />
        </div>
        <div class="mt-1 text-xs text-g-600">5 段格式：分 时 日 月 周，最小粒度到分钟</div>
      </ElFormItem>

      <ElFormItem label="状态" prop="status">
        <ElSwitch v-model="form.status" :active-value="1" :inactive-value="0" />
        <span class="ml-2 text-xs text-g-600">{{
          form.status === 1 ? '启用，按频率调度' : '停用，不会被调度'
        }}</span>
      </ElFormItem>

      <ElFormItem label="备注" prop="remark">
        <ElInput v-model="form.remark" type="textarea" :rows="2" maxlength="255" show-word-limit />
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
  import { fetchCreateJob, fetchUpdateJob } from '@/api/infra/job'

  defineOptions({ name: 'JobFormDialog' })

  type JobItem = Api.Infra.JobItem
  type HandlerItem = Api.Infra.JobHandlerItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    jobData?: JobItem | null
    handlers: HandlerItem[]
  }

  const props = withDefaults(defineProps<Props>(), { jobData: null })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const CRON_PRESETS = [
    { label: '每分钟', value: '* * * * *' },
    { label: '每 5 分钟', value: '*/5 * * * *' },
    { label: '每 10 分钟', value: '*/10 * * * *' },
    { label: '每小时', value: '0 * * * *' },
    { label: '每天 03:00', value: '0 3 * * *' },
    { label: '每周一 02:00', value: '0 2 * * 1' },
    { label: '每月 1 日 02:00', value: '0 2 1 * *' }
  ]

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const cronPreset = ref('')

  const form = reactive({
    handler: '',
    name: '',
    cron: '',
    status: 1,
    remark: ''
  })

  const currentHandler = computed(
    () => props.handlers.find((item) => item.key === form.handler) || null
  )

  const rules: FormRules = {
    handler: [{ required: true, message: '请选择任务处理器', trigger: 'change' }],
    name: [
      { required: true, message: '请输入任务名称', trigger: 'blur' },
      { max: 64, message: '任务名称不能超过 64 个字符', trigger: 'blur' }
    ],
    cron: [
      { required: true, message: '请填写执行频率', trigger: 'blur' },
      { max: 64, message: 'cron 表达式不能超过 64 个字符', trigger: 'blur' }
    ]
  }

  // 选择处理器时带出默认名称与默认频率，用户已填过则不覆盖
  const handleHandlerChange = (key: string): void => {
    const def = props.handlers.find((item) => item.key === key)
    if (!def) return

    if (!form.name.trim()) form.name = def.name
    if (!form.cron.trim()) form.cron = def.defaultCron
  }

  const handlePreset = (value: string): void => {
    if (value) form.cron = value
  }

  // 打开弹窗时回填；新增时清空并重置校验
  watch(
    () => [props.visible, props.jobData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.jobData
      Object.assign(form, {
        handler: data?.handler ?? '',
        name: data?.name ?? '',
        cron: data?.cron ?? '',
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      cronPreset.value = CRON_PRESETS.some((item) => item.value === form.cron) ? form.cron : ''

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
      const payload: Api.Infra.JobSaveParams = {
        name: form.name.trim(),
        handler: form.handler,
        cron: form.cron.trim(),
        status: form.status,
        remark: form.remark.trim()
      }

      if (isEdit.value && props.jobData) {
        await fetchUpdateJob(props.jobData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateJob(payload)
        ElMessage.success('新增成功')
      }

      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出（cron 不合法等由后端返回）
    } finally {
      submitting.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .job-alert {
    margin-bottom: 18px;
    border-radius: 8px;
  }

  .job-form {
    :deep(.el-form-item) {
      margin-bottom: 18px;
    }

    :deep(.el-form-item__label) {
      padding-bottom: 4px;
      font-weight: 500;
    }
  }
</style>
