<!-- 基础设施 / 代码生成：选表 → 配置模块 → 预览 → 落盘 -->
<template>
  <div class="infra-generator-page art-full-height">
    <!-- 配置条 -->
    <ElCard class="gen-config" shadow="never">
      <div class="config-grid">
        <div class="config-item">
          <label class="config-label">数据表</label>
          <ElSelect
            v-model="form.table"
            filterable
            placeholder="选择要生成的表"
            class="w-full"
            :loading="tableLoading"
            @change="handleTableChange"
          >
            <ElOption v-for="item in tables" :key="item.name" :label="item.name" :value="item.name">
              <div class="option-row">
                <span class="option-name">{{ item.name }}</span>
                <span class="option-comment">{{ item.comment || '无注释' }}</span>
              </div>
            </ElOption>
          </ElSelect>
        </div>

        <div class="config-item">
          <label class="config-label">目标模块</label>
          <ElInput v-model="form.module" placeholder="如 demo / pay" />
        </div>

        <div class="config-item">
          <label class="config-label">结构体名前缀</label>
          <ElInput v-model="form.pascal" placeholder="默认按表名推导" />
        </div>

        <div class="config-item is-switch">
          <label class="config-label">覆盖已有文件</label>
          <ElSwitch v-model="form.overwrite" />
        </div>

        <div class="config-actions">
          <ElButton
            type="primary"
            :loading="previewing"
            :disabled="!form.table"
            @click="handlePreview"
            v-ripple
          >
            预览代码
          </ElButton>
          <ElButton
            v-auth="'infra:generator:write'"
            :loading="generating"
            :disabled="!form.table || !form.module"
            @click="handleGenerate"
            v-ripple
          >
            生成落盘
          </ElButton>
        </div>
      </div>

      <div class="config-hint">
        <ArtSvgIcon icon="ri:information-line" class="hint-icon" />
        <span>
          生成 model / repository / service / request / controller 与前端
          api、页面；路由片段只预览。 已存在文件默认跳过，开启覆盖才会重写。
        </span>
      </div>
    </ElCard>

    <!-- 左：表结构 / 右：生成结果 -->
    <div class="gen-main">
      <ElCard class="gen-pane" shadow="never">
        <div class="pane-head">
          <div class="pane-title">
            表结构
            <span v-if="selectedTable" class="pane-title-extra">{{ selectedTable.name }}</span>
          </div>
          <div class="pane-meta">
            <template v-if="columns.length">
              共 {{ columns.length }} 字段 · 业务 {{ businessCount }} 项
            </template>
            <template v-else>请选择数据表</template>
          </div>
        </div>

        <ElScrollbar v-if="columns.length" class="pane-body">
          <div class="field-list">
            <div class="field-row is-head">
              <span>字段</span>
              <span>类型</span>
              <span>Go / 表单</span>
            </div>

            <div
              v-for="col in columns"
              :key="col.name"
              class="field-row"
              :class="{ 'is-audit': auditColumns.includes(col.name) }"
            >
              <div class="field-name">
                <span class="mono">{{ col.name }}</span>
                <ElTag v-if="col.primaryKey" size="small" type="warning" effect="plain">PK</ElTag>
                <ElTag
                  v-if="auditColumns.includes(col.name)"
                  size="small"
                  type="info"
                  effect="plain"
                >
                  审计
                </ElTag>
                <div v-if="col.comment" class="field-comment">{{ col.comment }}</div>
              </div>

              <div class="field-type">
                <span class="mono">{{ col.columnType }}</span>
                <div v-if="col.nullable" class="field-comment">可空</div>
              </div>

              <div class="field-form">
                <span class="mono field-go-type">{{ col.goType }}</span>
                <ElTag size="small" type="info" effect="plain">{{ col.formType }}</ElTag>
              </div>
            </div>
          </div>
        </ElScrollbar>

        <ElEmpty v-else description="选择表后展示字段" :image-size="80" class="pane-empty" />
      </ElCard>

      <ElCard class="gen-pane" shadow="never">
        <div class="pane-head">
          <div class="pane-title">
            生成结果
            <span v-if="files.length" class="pane-title-extra">
              {{ files.length }} 个文件 · {{ writableCount }} 个可落盘
            </span>
          </div>
          <div class="pane-head-actions">
            <ElButton v-if="files.length" link type="primary" @click="handlePreview">
              重新预览
            </ElButton>
            <ElButton v-if="activeFile" link type="primary" @click="handleCopy"
              >复制当前文件</ElButton
            >
          </div>
        </div>

        <div v-if="files.length" class="pane-body is-code">
          <ElTabs v-model="activeFile" class="file-tabs">
            <ElTabPane v-for="file in files" :key="file.path" :name="file.path">
              <template #label>
                <span class="tab-label">{{ fileName(file.path) }}</span>
                <span v-if="!file.writable" class="tab-badge">预览</span>
              </template>
            </ElTabPane>
          </ElTabs>

          <div class="code-meta">
            <span class="mono code-path">{{ activeFile }}</span>
            <span class="code-stats">{{ languageOf(activeFile) }} · {{ lineCount }} 行</span>
            <ElTag v-if="activeFileWritable" size="small" type="success" effect="plain"
              >可落盘</ElTag
            >
            <ElTag v-else size="small" type="info" effect="plain">仅预览</ElTag>
          </div>

          <ElScrollbar class="code-scroll">
            <div class="editor">
              <div class="editor-gutter" aria-hidden="true">
                <span v-for="line in lineCount" :key="line">{{ line }}</span>
              </div>
              <pre
                class="editor-code"
              ><code class="gen-code-view" v-html="highlightedCode"></code></pre>
            </div>
          </ElScrollbar>
        </div>

        <div v-else class="pane-empty">
          <ElEmpty description="还没有生成内容" :image-size="90">
            <div class="empty-steps">
              <div>1. 选择一张表，自动带出模块与结构体名</div>
              <div>2. 点「预览代码」核对产物</div>
              <div>3. 确认无误后「生成落盘」</div>
            </div>
          </ElEmpty>
        </div>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import {
    ElButton,
    ElCard,
    ElEmpty,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElScrollbar,
    ElSelect,
    ElSwitch,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import hljs from 'highlight.js'
  import {
    fetchGenerateCode,
    fetchGetGeneratorColumns,
    fetchGetGeneratorTables,
    fetchPreviewGeneratorCode
  } from '@/api/infra/generator'

  defineOptions({ name: 'InfraGenerator' })

  type GeneratorFile = Api.Infra.GeneratorFile

  // 框架统一维护的审计字段，列表里灰显，避免用户以为它们要单独填
  const auditColumns = ['creator', 'created_at', 'updater', 'updated_at', 'deleted_at']

  const tables = ref<Api.Infra.GeneratorTable[]>([])
  const columns = ref<Api.Infra.GeneratorColumn[]>([])
  const files = ref<GeneratorFile[]>([])
  const activeFile = ref('')
  const tableLoading = ref(false)
  const previewing = ref(false)
  const generating = ref(false)

  const form = reactive({
    table: '',
    module: '',
    pascal: '',
    overwrite: false
  })

  // 记录自动填充值：用户手动改过之后就不再覆盖
  const autoModule = ref('')
  const autoPascal = ref('')

  const selectedTable = computed(
    () => tables.value.find((item) => item.name === form.table) || null
  )
  const businessCount = computed(
    () => columns.value.filter((col) => !col.primaryKey && !auditColumns.includes(col.name)).length
  )
  const writableCount = computed(() => files.value.filter((file) => file.writable).length)
  const activeFileContent = computed(
    () => files.value.find((file) => file.path === activeFile.value)?.content ?? ''
  )
  const activeFileWritable = computed(
    () => files.value.find((file) => file.path === activeFile.value)?.writable ?? false
  )

  const fileName = (path: string): string => path.split('/').slice(-1)[0] || path

  // 按产物后缀选高亮语言：Vue 用 xml（模板 + 脚本足够可读），其余纯文本不高亮
  const languageOf = (path: string): string => {
    if (path.endsWith('.go')) return 'go'
    if (path.endsWith('.ts')) return 'typescript'
    if (path.endsWith('.vue')) return 'xml'
    return 'plaintext'
  }

  const lineCount = computed(() => {
    const code = activeFileContent.value
    return code ? code.split('\n').length : 0
  })

  // hljs.highlight 自带转义，失败时退回纯文本，避免预览白屏
  const highlightedCode = computed(() => {
    const code = activeFileContent.value
    if (!code) return ''

    const language = languageOf(activeFile.value)
    if (language === 'plaintext') return code.replace(/&/g, '&amp;').replace(/</g, '&lt;')

    try {
      return hljs.highlight(code, { language }).value
    } catch {
      return code.replace(/&/g, '&amp;').replace(/</g, '&lt;')
    }
  })

  // member_user → MemberUser
  const toPascal = (value: string): string =>
    value
      .split(/[_-]/)
      .filter(Boolean)
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join('')

  const loadTables = async (): Promise<void> => {
    tableLoading.value = true
    try {
      tables.value = (await fetchGetGeneratorTables()) ?? []
    } catch {
      tables.value = []
    } finally {
      tableLoading.value = false
    }
  }

  const handleTableChange = async (table: string): Promise<void> => {
    files.value = []
    columns.value = []
    activeFile.value = ''

    // 默认模块取表名前缀（member_user → member）
    const prefix = table.split('_')[0] || ''
    if (prefix && (!form.module || form.module === autoModule.value)) {
      form.module = prefix
      autoModule.value = prefix
    }

    // 结构体名与后端保持同一规则：去掉末尾单个 s 再转大驼峰（pay_orders → PayOrder）
    const pascal = toPascal(table.endsWith('s') ? table.slice(0, -1) : table)
    if (pascal && (!form.pascal || form.pascal === autoPascal.value)) {
      form.pascal = pascal
      autoPascal.value = pascal
    }

    try {
      columns.value = (await fetchGetGeneratorColumns(table)) ?? []
    } catch {
      columns.value = []
    }
  }

  const handlePreview = async (): Promise<void> => {
    if (!form.module) {
      ElMessage.warning('请填写目标模块')
      return
    }

    previewing.value = true
    try {
      const res = await fetchPreviewGeneratorCode({
        table: form.table,
        module: form.module.trim(),
        pascal: form.pascal.trim() || undefined,
        overwrite: form.overwrite
      })
      files.value = res ?? []
      activeFile.value = files.value[0]?.path ?? ''
      if (!files.value.length) ElMessage.warning('没有生成内容')
    } catch {
      files.value = []
    } finally {
      previewing.value = false
    }
  }

  // 落盘前二次确认：会写真实文件，覆盖模式下风险更高
  const handleGenerate = async (): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        form.overwrite
          ? `将在 app/modules/${form.module} 与 ui/src 下生成文件，并覆盖同名文件。确认继续？`
          : `将在 app/modules/${form.module} 与 ui/src 下生成文件，已存在的文件会跳过。确认继续？`,
        '生成代码',
        { confirmButtonText: '生成', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    generating.value = true
    try {
      const res = await fetchGenerateCode({
        table: form.table,
        module: form.module.trim(),
        pascal: form.pascal.trim() || undefined,
        overwrite: form.overwrite
      })

      const written = res?.written?.length ?? 0
      const skipped = res?.skipped ?? []
      if (skipped.length > 0) {
        ElMessageBox.alert(
          `已写入 ${written} 个文件；以下 ${skipped.length} 个已存在被跳过：\n${skipped.join('\n')}`,
          '生成完成',
          { confirmButtonText: '知道了' }
        )
      } else {
        ElMessage.success(`生成完成，写入 ${written} 个文件`)
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      generating.value = false
    }
  }

  // 复制当前文件内容
  const handleCopy = async (): Promise<void> => {
    if (!activeFileContent.value) return
    try {
      await navigator.clipboard.writeText(activeFileContent.value)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动选中复制')
    }
  }

  onMounted(loadTables)
</script>

<style lang="scss" scoped>
  .infra-generator-page {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 0;
  }

  .gen-config {
    flex: none;

    :deep(.el-card__body) {
      padding: 16px 20px 12px;
    }
  }

  .config-grid {
    display: grid;
    gap: 12px 16px;
    align-items: end;
    grid-template-columns: minmax(220px, 1.4fr) 1fr 1fr 120px auto;
  }

  .config-item {
    min-width: 0;

    &.is-switch {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 32px;
    }
  }

  .config-label {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
  }

  .config-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  .config-hint {
    display: flex;
    gap: 6px;
    align-items: flex-start;
    padding-top: 10px;
    margin-top: 12px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    border-top: 1px dashed var(--el-border-color-lighter);
  }

  .hint-icon {
    flex: none;
    margin-top: 2px;
    font-size: 14px;
    color: var(--el-color-primary);
  }

  .option-row {
    display: flex;
    gap: 8px;
    align-items: baseline;
    justify-content: space-between;
  }

  .option-name {
    font-family: var(--el-font-family-monospace, monospace);
  }

  .option-comment {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  // 左右分栏：整页高度内部分滚，不留大片空白
  .gen-main {
    display: grid;
    flex: 1;
    gap: 12px;
    min-height: 360px;
    grid-template-columns: 380px minmax(0, 1fr);

    @media (max-width: 1100px) {
      grid-template-columns: 1fr;
    }
  }

  .gen-pane {
    display: flex;
    flex-direction: column;
    min-height: 0;

    :deep(.el-card__body) {
      display: flex;
      flex: 1;
      flex-direction: column;
      min-height: 0;
      padding: 0;
      overflow: hidden;
    }
  }

  .pane-head {
    display: flex;
    flex: none;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .pane-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .pane-title-extra {
    margin-left: 6px;
    font-size: 12px;
    font-weight: 400;
    color: var(--el-text-color-secondary);
  }

  .pane-meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .pane-head-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .pane-body {
    flex: 1;
    min-height: 0;
  }

  .pane-empty {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .empty-steps {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 4px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  // 字段列表：表头固定，内容滚动
  .field-list {
    padding: 0 4px 8px;
  }

  .field-row {
    display: grid;
    gap: 8px;
    align-items: center;
    padding: 8px 12px;
    font-size: 12px;
    grid-template-columns: minmax(0, 1.3fr) minmax(0, 0.9fr) 92px;
    border-bottom: 1px solid var(--el-border-color-lighter);

    &.is-head {
      position: sticky;
      top: 0;
      z-index: 1;
      font-weight: 500;
      color: var(--el-text-color-secondary);
      background: var(--el-fill-color-lighter);
    }

    &.is-audit {
      background: var(--el-fill-color-lighter);
    }
  }

  .field-name {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    min-width: 0;
  }

  .field-type,
  .field-form {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    min-width: 0;
  }

  .field-go-type {
    font-size: 11px;
    color: var(--el-text-color-secondary);
  }

  .field-comment {
    width: 100%;
    font-size: 11px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
    word-break: break-all;
  }

  .mono {
    font-family: var(--el-font-family-monospace, monospace);
  }

  // 代码预览：文件 tab + 路径 + 内容各自占位，内容内部滚动
  .pane-body.is-code {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .file-tabs {
    flex: none;
    padding: 0 12px;

    :deep(.el-tabs__header) {
      margin-bottom: 0;
    }

    :deep(.el-tabs__item) {
      height: 36px;
      padding: 0 12px;
      font-size: 12px;
    }
  }

  .tab-label {
    font-family: var(--el-font-family-monospace, monospace);
  }

  .tab-badge {
    padding: 0 4px;
    margin-left: 4px;
    font-size: 10px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-dark);
    border-radius: 4px;
  }

  .code-meta {
    display: flex;
    flex: none;
    gap: 8px;
    align-items: center;
    padding: 8px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .code-path {
    flex: 1;
    overflow: hidden;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .code-stats {
    flex: none;
    font-size: 11px;
    color: var(--el-text-color-placeholder);
  }

  .code-scroll {
    flex: 1;
    min-height: 0;
    background: #282c34;
  }

  // 只读代码视图：左侧行号吸顶，右侧代码横向滚动，行高统一 20px 保证对齐
  .editor {
    display: flex;
    min-width: min-content;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    line-height: 20px;
    color: #a6accd;
    background: #282c34;
  }

  .editor-gutter {
    position: sticky;
    left: 0;
    z-index: 1;
    flex: none;
    padding: 12px 10px 12px 12px;
    color: #5c6370;
    text-align: right;
    user-select: none;
    background: #21252b;
    border-right: 1px solid rgb(255 255 255 / 6%);

    span {
      display: block;
      height: 20px;
    }
  }

  .editor-code {
    flex: 1;
    padding: 12px 16px;
    margin: 0;
    line-height: 20px;
    white-space: pre;
    tab-size: 2;
  }
</style>

<style lang="scss">
  // highlight.js 的 token 配色（one-dark 风格），用 .gen-code-view 限定作用域，避免污染全局
  .gen-code-view {
    display: block;

    .hljs-comment,
    .hljs-quote {
      font-style: italic;
      color: #6f747d;
    }

    .hljs-keyword,
    .hljs-formula,
    .hljs-doctag {
      color: #c792ea;
    }

    .hljs-string,
    .hljs-section,
    .hljs-selector-class,
    .hljs-template-variable,
    .hljs-deletion {
      color: #aed07e;
    }

    .hljs-name,
    .hljs-selector-tag,
    .hljs-subst {
      color: #c86068;
    }

    .hljs-literal,
    .hljs-number {
      color: #56b6c2;
    }

    .hljs-title,
    .hljs-function .hljs-title {
      color: #82aaff;
    }

    .hljs-type,
    .hljs-built_in,
    .hljs-class .hljs-title {
      color: #ffcb6b;
    }

    .hljs-attr,
    .hljs-attribute,
    .hljs-variable,
    .hljs-template-tag {
      color: #f78c6c;
    }

    .hljs-meta,
    .hljs-symbol,
    .hljs-bullet {
      color: #89ddff;
    }
  }
</style>
