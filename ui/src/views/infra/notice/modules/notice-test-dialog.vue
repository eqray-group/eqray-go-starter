<!-- 测试发送：邮件填收件人，机器人直接确认；结果就地展示，不弹两次 toast -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="测试发送"
    width="520px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="test-head">
      <ElTag :type="typeTag" size="small">{{ typeName }}</ElTag>
      <span class="test-name">{{ channel?.name }}</span>
    </div>

    <ElForm label-position="top">
      <ElFormItem v-if="channel?.type === 'email'" label="收件人">
        <ElInput v-model="target" placeholder="name@example.com" clearable />
      </ElFormItem>
      <div v-else class="test-hint">
        将向该{{ channel?.type === 'wecom' ? '企业微信群' : '钉钉群' }}发送一条内置模板的测试消息，
        用于验证 Webhook 与签名是否配置正确。
      </div>
    </ElForm>

    <div v-if="result" class="test-result" :class="result.status === 1 ? 'is-ok' : 'is-fail'">
      <div class="result-line">
        <ArtSvgIcon
          :icon="result.status === 1 ? 'ri:checkbox-circle-line' : 'ri:close-circle-line'"
          class="result-icon"
        />
        <span>{{ result.status === 1 ? '发送成功' : '发送失败' }}</span>
        <span class="result-duration">{{ result.duration }} ms</span>
      </div>
      <pre v-if="result.error" class="result-error">{{ result.error }}</pre>
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">关闭</ElButton>
      <ElButton type="primary" :loading="sending" @click="handleSend">
        {{ result ? '再次发送' : '发送' }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElButton, ElDialog, ElForm, ElFormItem, ElInput, ElMessage, ElTag } from 'element-plus'
  import { fetchTestNoticeChannel } from '@/api/infra/notice'

  defineOptions({ name: 'NoticeTestDialog' })

  const props = defineProps<{
    visible: boolean
    channel: Api.Infra.NoticeChannelItem | null
    typeName: string
  }>()

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const typeTag = computed(() => {
    switch (props.channel?.type) {
      case 'email':
        return 'primary'
      case 'wecom':
        return 'success'
      default:
        return 'warning'
    }
  })

  const target = ref('')
  const sending = ref(false)
  const result = ref<Api.Infra.NoticeLogItem | null>(null)

  // 每次打开重置上一次的结果，避免误以为这次已发送
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      target.value = ''
      result.value = null
    }
  )

  const handleSend = async (): Promise<void> => {
    if (!props.channel) return
    if (props.channel.type === 'email' && !target.value.trim()) {
      ElMessage.warning('请填写收件人')
      return
    }

    sending.value = true
    try {
      const log = await fetchTestNoticeChannel(props.channel.id, target.value.trim() || undefined)
      result.value = log ?? null
      if (log?.status === 1) {
        ElMessage.success('发送成功')
      } else {
        ElMessage.error('发送失败，请查看下方原因')
      }
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      sending.value = false
    }
  }
</script>

<style lang="scss" scoped>
  .test-head {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 12px;
  }

  .test-hint {
    padding: 10px 12px;
    font-size: 12px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
  }

  .test-name {
    font-size: 13px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }

  .test-result {
    padding: 10px 12px;
    margin-top: 4px;
    border-radius: 8px;

    &.is-ok {
      background: var(--el-color-success-light-9);
      border: 1px solid var(--el-color-success-light-7);
    }

    &.is-fail {
      background: var(--el-color-danger-light-9);
      border: 1px solid var(--el-color-danger-light-7);
    }
  }

  .result-line {
    display: flex;
    gap: 6px;
    align-items: center;
    font-size: 13px;
    font-weight: 500;
  }

  .result-icon {
    font-size: 16px;
  }

  .is-ok .result-icon,
  .is-ok .result-line {
    color: var(--el-color-success);
  }

  .is-fail .result-icon,
  .is-fail .result-line {
    color: var(--el-color-danger);
  }

  .result-duration {
    margin-left: auto;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    font-weight: 400;
    color: var(--el-text-color-secondary);
  }

  .result-error {
    max-height: 180px;
    padding: 8px 10px;
    margin: 8px 0 0;
    overflow: auto;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    background: rgb(255 255 255 / 60%);
    border-radius: 6px;
  }
</style>
