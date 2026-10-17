<template>
  <ElDialog
    v-model="dialogVisible"
    :title="`分配角色 - ${user?.username || ''}`"
    width="480px"
    align-center
    :close-on-click-modal="false"
  >
    <div v-loading="loading" class="role-list">
      <ElCheckboxGroup v-model="checkedRoleIDs">
        <ElCheckbox v-for="role in roleOptions" :key="role.id" :value="role.id" class="role-item">
          {{ role.name }}
          <span class="role-code">{{ role.code }}</span>
        </ElCheckbox>
      </ElCheckboxGroup>
      <ElEmpty v-if="!loading && roleOptions.length === 0" description="暂无可分配的角色" />
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchAssignUserRoles, fetchGetUserRoles } from '@/api/system/user'
  import { fetchGetRoleOptions } from '@/api/system/role'

  type UserItem = Api.SystemManage.UserListItem
  type RoleItem = Api.SystemManage.RoleItem

  interface Props {
    visible: boolean
    user?: Partial<UserItem> | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), { user: null })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const loading = ref(false)
  const submitting = ref(false)
  const roleOptions = ref<RoleItem[]>([])
  const checkedRoleIDs = ref<string[]>([])

  const loadData = async (): Promise<void> => {
    if (!props.user?.id) return

    loading.value = true
    try {
      const [roles, granted] = await Promise.all([
        fetchGetRoleOptions(),
        fetchGetUserRoles(props.user.id)
      ])
      roleOptions.value = roles
      checkedRoleIDs.value = granted
    } finally {
      loading.value = false
    }
  }

  watch(
    () => props.visible,
    (visible) => {
      if (visible) loadData()
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!props.user?.id) return

    submitting.value = true
    try {
      await fetchAssignUserRoles(props.user.id, checkedRoleIDs.value)
      ElMessage.success('分配成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .role-list {
    max-height: 360px;
    overflow: auto;
  }

  .role-item {
    display: flex;
    width: 100%;
    margin-right: 0;
  }

  .role-code {
    margin-left: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
</style>
