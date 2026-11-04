<!-- 支付联调：选会员 + 渠道（可全部），批量建 1 分钱测试单并展示拉起参数 -->
<template>
  <div class="pay-test-page art-full-height">
    <ElCard class="mb-4">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium">联调下单</span>
          <ElTag type="info" effect="plain"
            >金额固定 0.01 元 · 写入 pay_order（biz_type = pay_test）</ElTag
          >
        </div>
      </template>

      <ElAlert
        type="warning"
        :closable="false"
        show-icon
        class="test-alert"
        title="测试单会真实走支付流程"
      >
        <span class="text-xs">
          钱包渠道会真实扣减会员余额（每个渠道 1 分）；模拟渠道需在参数配置把 pay.mock.enabled 置为
          1； 微信 / 支付宝需要真实商户凭据与公网 HTTPS 的 APP_URL。
        </span>
      </ElAlert>

      <ElForm :inline="true" class="test-form">
        <ElFormItem label="会员">
          <div class="flex items-center gap-2">
            <ElInput
              :model-value="memberLabel"
              readonly
              placeholder="请选择付款会员"
              style="width: 300px"
            >
              <template #append>
                <ElButton @click="memberPickerVisible = true">选择会员</ElButton>
              </template>
            </ElInput>
          </div>
        </ElFormItem>

        <ElFormItem label="支付渠道">
          <ElSelect
            v-model="form.channelCodes"
            multiple
            collapse-tags
            collapse-tags-tooltip
            placeholder="不选则测试全部启用渠道"
            style="width: 320px"
          >
            <ElOption
              v-for="item in channelOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem>
          <ElSpace>
            <ElButton type="primary" :loading="creating" @click="handleCreate()"
              >创建测试单</ElButton
            >
            <ElButton :loading="creating" @click="handleCreate(true)">一键测试全部渠道</ElButton>
            <ElButton :loading="refreshingAll" :disabled="!rows.length" @click="handleRefreshAll">
              刷新全部
            </ElButton>
          </ElSpace>
        </ElFormItem>
      </ElForm>

      <div v-if="!channelOptions.length" class="mt-2 text-xs text-g-600">
        当前没有启用中的支付渠道，请先到「支付管理 → 支付渠道」新增并启用。
      </div>
    </ElCard>

    <ElCard v-if="rows.length">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-medium">测试结果（{{ rows.length }} 个渠道）</span>
          <span class="text-xs text-g-600">展开行查看二维码 / 跳转链接 / 调起参数</span>
        </div>
      </template>

      <ElTable :data="rows" row-key="channelCode" border>
        <ElTableColumn type="expand">
          <template #default="{ row }">
            <div class="pay-info">
              <ElAlert
                v-if="row.error"
                type="error"
                :closable="false"
                show-icon
                :title="row.error"
              />
              <template v-else-if="row.payInfo">
                <!-- 二维码 -->
                <div v-if="row.payInfo.kind === 'qr'" class="flex items-center gap-5">
                  <QrcodeVue
                    v-if="row.payInfo.url"
                    :value="row.payInfo.url"
                    :size="160"
                    level="M"
                  />
                  <div class="flex flex-col gap-2">
                    <span class="text-xs text-g-600">用手机扫码支付，付完点「刷新」</span>
                    <ElInput :model-value="row.payInfo.url" readonly style="width: 420px" />
                    <div>
                      <ElButton size="small" @click="copyText(row.payInfo.url)">复制链接</ElButton>
                    </div>
                  </div>
                </div>

                <!-- 跳转 -->
                <div v-else-if="row.payInfo.kind === 'redirect'" class="flex items-center gap-3">
                  <ElInput :model-value="row.payInfo.url" readonly style="width: 520px" />
                  <ElButton size="small" type="primary" plain @click="openUrl(row)"
                    >打开支付页</ElButton
                  >
                </div>

                <!-- 余额支付 -->
                <ElAlert
                  v-else-if="row.payInfo.kind === 'wallet'"
                  type="success"
                  :closable="false"
                  show-icon
                  title="余额支付已完成，无需继续操作"
                />

                <!-- 模拟支付 -->
                <div v-else-if="row.payInfo.kind === 'mock'" class="flex items-center gap-3">
                  <span class="text-xs text-g-600">
                    模拟渠道：点击行内「模拟支付」完成付款（需 pay.mock.enabled = 1）
                  </span>
                </div>

                <!-- JSAPI / 小程序 / App -->
                <pre v-else-if="row.payInfo.params" class="pay-params">{{
                  JSON.stringify(row.payInfo.params, null, 2)
                }}</pre>

                <span v-else class="text-xs text-g-600">该渠道没有返回可展示的支付参数</span>
              </template>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="支付渠道" min-width="150">
          <template #default="{ row }">
            <div class="flex flex-col">
              <span>{{ channelLabels[row.channelCode] || row.channelCode }}</span>
              <span class="text-xs text-g-600">{{ row.channelCode }}</span>
            </div>
          </template>
        </ElTableColumn>

        <ElTableColumn label="支付单号" min-width="200">
          <template #default="{ row }">
            <div v-if="row.merchantOrderNo" class="flex items-center gap-2">
              <span class="font-mono text-xs">{{ row.merchantOrderNo }}</span>
              <ElButton link type="primary" @click="copyText(row.merchantOrderNo)">复制</ElButton>
            </div>
            <span v-else class="text-g-600">-</span>
          </template>
        </ElTableColumn>

        <ElTableColumn label="金额" width="100">
          <template #default="{ row }">{{ formatPrice(row.price) }}</template>
        </ElTableColumn>

        <ElTableColumn label="状态" width="110">
          <template #default="{ row }">
            <ElTag v-if="row.error" type="danger" size="small">失败</ElTag>
            <ElTag v-else :type="STATUS_TAG[row.status] || 'info'" size="small">
              {{ STATUS_TEXT[row.status] || '未知' }}
            </ElTag>
          </template>
        </ElTableColumn>

        <ElTableColumn label="操作" width="230" fixed="right">
          <template #default="{ row }">
            <div v-if="!row.error" class="flex items-center">
              <ElButton
                link
                type="primary"
                :loading="busy[row.channelCode]"
                :disabled="row.status === 20 || row.status === 30"
                @click="handleRefreshRow(row)"
              >
                刷新
              </ElButton>
              <ElButton
                link
                type="success"
                :loading="busy[row.channelCode]"
                :disabled="row.status === 20 || row.status === 30"
                @click="handleSync(row)"
              >
                查单
              </ElButton>
              <ElButton
                v-if="row.payInfo?.kind === 'mock'"
                link
                type="warning"
                :loading="busy[row.channelCode]"
                :disabled="row.status === 20"
                @click="handleMockPay(row)"
              >
                模拟支付
              </ElButton>
              <ElButton v-if="row.payInfo?.kind === 'redirect'" link @click="openUrl(row)">
                打开
              </ElButton>
            </div>
            <span v-else class="text-xs text-danger">见展开行原因</span>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <MemberPickerDialog v-model:visible="memberPickerVisible" @select="handleSelectMember" />
  </div>
</template>

<script setup lang="ts">
  import QrcodeVue from 'qrcode.vue'
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElForm,
    ElFormItem,
    ElInput,
    ElMessage,
    ElOption,
    ElSelect,
    ElSpace,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import { fetchGetPayChannelList } from '@/api/pay/channel'
  import {
    fetchCreatePayTestOrdersBatch,
    fetchGetPayTestOrder,
    fetchMockPayTestOrder,
    fetchSyncPayTestOrder
  } from '@/api/pay/test'
  import MemberPickerDialog from './modules/member-picker-dialog.vue'

  defineOptions({ name: 'PayTest' })

  type TestRow = Api.Pay.TestOrderOutcome
  type MemberItem = Api.Member.MemberItem

  const STATUS_TEXT: Record<number, string> = {
    0: '待支付',
    10: '支付中',
    20: '已支付',
    30: '已关闭'
  }
  const STATUS_TAG: Record<number, 'success' | 'warning' | 'info' | 'danger'> = {
    0: 'warning',
    10: 'warning',
    20: 'success',
    30: 'info'
  }

  const form = reactive({
    channelCodes: [] as string[]
  })

  const member = ref<MemberItem | null>(null)
  const memberPickerVisible = ref(false)
  const channelOptions = ref<{ label: string; value: string }[]>([])
  const channelLabels = ref<Record<string, string>>({})
  const rows = ref<TestRow[]>([])
  const creating = ref(false)
  const refreshingAll = ref(false)
  // 行内按钮的 loading，按渠道编码区分
  const busy = reactive<Record<string, boolean>>({})

  const memberLabel = computed(() =>
    member.value
      ? `${member.value.nickname || '-'}（${member.value.mobile || '-'}）· ${member.value.id}`
      : ''
  )

  const formatPrice = (price?: number): string => `${((price ?? 0) / 100).toFixed(2)} 元`

  // 只展示启用中的渠道，名称后带编码便于对照
  const loadChannels = async (): Promise<void> => {
    try {
      const res = await fetchGetPayChannelList({ status: 1, pageNo: 1, pageSize: 100 })
      const list = res?.list ?? []
      channelOptions.value = list.map((item) => ({
        label: `${item.name}（${item.code}）`,
        value: item.code
      }))
      channelLabels.value = Object.fromEntries(list.map((item) => [item.code, item.name]))
    } catch {
      channelOptions.value = []
      channelLabels.value = {}
    }
  }

  const handleSelectMember = (picked: MemberItem): void => {
    member.value = picked
  }

  // all=true 时忽略已选渠道，覆盖全部启用渠道
  const handleCreate = async (all = false): Promise<void> => {
    if (!member.value) {
      ElMessage.warning('请先选择付款会员')

      return
    }
    if (all && !channelOptions.value.length) {
      ElMessage.warning('当前没有启用中的支付渠道')

      return
    }

    creating.value = true
    try {
      const outcomes = await fetchCreatePayTestOrdersBatch({
        userId: member.value.id,
        channelCodes: all ? [] : form.channelCodes
      })
      rows.value = outcomes ?? []

      const failed = rows.value.filter((row) => row.error).length
      if (failed === 0) {
        ElMessage.success(`已创建 ${rows.value.length} 个测试单`)
      } else {
        ElMessage.warning(`${rows.value.length} 个渠道中有 ${failed} 个失败，展开行查看原因`)
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      creating.value = false
    }
  }

  // 单行刷新：只更新状态
  const handleRefreshRow = async (row: TestRow): Promise<void> => {
    if (!row.merchantOrderNo) return

    busy[row.channelCode] = true
    try {
      const order = await fetchGetPayTestOrder(row.merchantOrderNo)
      row.status = order.status
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      busy[row.channelCode] = false
    }
  }

  // 主动查单：向渠道拉一次真实状态
  const handleSync = async (row: TestRow): Promise<void> => {
    if (!row.merchantOrderNo) return

    busy[row.channelCode] = true
    try {
      const order = await fetchSyncPayTestOrder(row.merchantOrderNo)
      row.status = order.status
      ElMessage.success('已向渠道查单')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      busy[row.channelCode] = false
    }
  }

  const handleMockPay = async (row: TestRow): Promise<void> => {
    if (!row.merchantOrderNo) return

    busy[row.channelCode] = true
    try {
      const order = await fetchMockPayTestOrder(row.merchantOrderNo)
      row.status = order.status
      ElMessage.success('模拟支付成功')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      busy[row.channelCode] = false
    }
  }

  // 批量刷新：只处理仍未完结且有单号的行
  const handleRefreshAll = async (): Promise<void> => {
    const pending = rows.value.filter(
      (row) => row.merchantOrderNo && row.status !== 20 && row.status !== 30
    )
    if (!pending.length) {
      ElMessage.info('没有需要刷新的测试单')

      return
    }

    refreshingAll.value = true
    try {
      for (const row of pending) {
        await handleRefreshRow(row)
      }
      ElMessage.success('已刷新')
    } finally {
      refreshingAll.value = false
    }
  }

  const openUrl = (row: TestRow): void => {
    const url = row.payInfo?.url
    if (url) window.open(url, '_blank')
  }

  const copyText = async (text?: string): Promise<void> => {
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动选中复制')
    }
  }

  onMounted(loadChannels)
</script>

<style lang="scss" scoped>
  // 卡片上下留白：提示条、表单不再贴着卡片边缘
  .pay-test-page {
    :deep(.el-card__body) {
      padding: 20px;
    }
  }

  .test-alert {
    margin-bottom: 20px;
    border-radius: 8px;

    :deep(.el-alert__content) {
      padding: 4px 0;
      line-height: 1.7;
    }
  }

  .test-form {
    :deep(.el-form-item) {
      margin-top: 2px;
      margin-bottom: 12px;
    }

    :deep(.el-form-item:last-child) {
      margin-bottom: 0;
    }
  }

  .pay-info {
    padding: 12px 24px;
  }

  .pay-params {
    max-height: 320px;
    padding: 12px 16px;
    overflow: auto;
    font-size: 12px;
    line-height: 1.7;
    border-radius: 8px;
    background: var(--el-fill-color-lighter);
  }
</style>
