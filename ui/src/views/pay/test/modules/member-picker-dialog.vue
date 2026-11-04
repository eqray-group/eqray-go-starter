<!-- 会员选择弹窗：联调时需要指定付款会员，避免手抄 ID 出错 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    title="选择会员"
    width="780px"
    align-center
    :close-on-click-modal="false"
  >
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <ElInput
        v-model="search.mobile"
        placeholder="手机号"
        clearable
        style="width: 180px"
        @keyup.enter="handleSearch"
      />
      <ElInput
        v-model="search.nickname"
        placeholder="昵称"
        clearable
        style="width: 180px"
        @keyup.enter="handleSearch"
      />
      <ElButton type="primary" @click="handleSearch">搜索</ElButton>
      <ElButton @click="handleReset">重置</ElButton>
    </div>

    <ElTable v-loading="loading" :data="list" height="380" border>
      <ElTableColumn prop="id" label="会员 ID" width="180" show-overflow-tooltip />
      <ElTableColumn label="会员" min-width="150">
        <template #default="{ row }">
          <div class="flex items-center gap-2">
            <ElAvatar :size="26" :src="row.avatar || undefined">
              {{ (row.nickname || '-').slice(0, 1) }}
            </ElAvatar>
            <span>{{ row.nickname || '-' }}</span>
          </div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="mobile" label="手机号" width="130" />
      <ElTableColumn label="状态" width="90">
        <template #default="{ row }">
          <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '禁用' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <ElButton link type="primary" @click="handlePick(row)">选择</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <div class="mt-3 flex justify-end">
      <ElPagination
        v-model:current-page="pageNo"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadMembers"
        @size-change="loadMembers"
      />
    </div>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    ElAvatar,
    ElButton,
    ElDialog,
    ElInput,
    ElPagination,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import { fetchGetMemberList } from '@/api/member/user'

  defineOptions({ name: 'MemberPickerDialog' })

  type MemberItem = Api.Member.MemberItem

  interface Props {
    visible: boolean
  }

  const props = defineProps<Props>()

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'select', member: MemberItem): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const list = ref<MemberItem[]>([])
  const total = ref(0)
  const pageNo = ref(1)
  const pageSize = ref(10)

  const search = reactive({
    mobile: '',
    nickname: ''
  })

  const loadMembers = async (): Promise<void> => {
    loading.value = true
    try {
      const res = await fetchGetMemberList({
        mobile: search.mobile.trim() || undefined,
        nickname: search.nickname.trim() || undefined,
        pageNo: pageNo.value,
        pageSize: pageSize.value
      })
      list.value = res?.list ?? []
      total.value = res?.total ?? 0
    } catch {
      // 错误提示已由请求层统一弹出（无 member:user:list 权限时会提示）
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  const handleSearch = (): void => {
    pageNo.value = 1
    loadMembers()
  }

  const handleReset = (): void => {
    search.mobile = ''
    search.nickname = ''
    handleSearch()
  }

  const handlePick = (member: MemberItem): void => {
    emit('select', member)
    dialogVisible.value = false
  }

  // 每次打开都回到第一页重新拉，避免看到过期数据
  watch(dialogVisible, (visible) => {
    if (!visible) return
    pageNo.value = 1
    loadMembers()
  })
</script>
