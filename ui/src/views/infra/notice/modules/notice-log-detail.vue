<!-- 发送记录展开行：失败原因与消息正文 -->
<template>
  <div class="notice-log-detail">
    <div class="log-meta">
      <span>模板：{{ row.templateCode || '—' }}</span>
      <span>接收目标：{{ row.target || '—' }}</span>
      <span>操作人：{{ row.creator || 'system' }}</span>
      <span>耗时：{{ row.duration }} ms</span>
    </div>

    <div v-if="row.error" class="log-block">
      <div class="log-label">失败原因</div>
      <pre class="log-pre is-error">{{ row.error }}</pre>
    </div>

    <div class="log-block">
      <div class="log-label">发送内容</div>
      <pre class="log-pre">{{ row.content || '（无内容）' }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'NoticeLogDetail' })

  defineProps<{
    row: Api.Infra.NoticeLogItem
  }>()
</script>

<style lang="scss" scoped>
  .notice-log-detail {
    padding: 4px 16px 12px 56px;
  }

  .log-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 24px;
    padding-bottom: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .log-block + .log-block {
    margin-top: 10px;
  }

  .log-label {
    margin-bottom: 4px;
    font-size: 12px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
  }

  .log-pre {
    max-height: 220px;
    padding: 10px 12px;
    margin: 0;
    overflow: auto;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--el-fill-color-lighter);
    border-radius: 6px;

    &.is-error {
      color: var(--el-color-danger);
      background: var(--el-color-danger-light-9);
    }
  }
</style>
