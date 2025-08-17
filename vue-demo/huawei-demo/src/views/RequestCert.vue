<script setup lang="ts">
import { ref, onMounted, h, resolveComponent, reactive } from 'vue';
import { message } from 'ant-design-vue';

type CertItem = {
  id: number
  publicKey: string
  ca: string
  status: 'pending' | 'issued' | 'rejected'
  description: string
}

const loading = ref(false)
const dataSource = ref<CertItem[]>([])


// ============ 列表列定义 ============
const columns = [
  { title: '公钥', dataIndex: 'publicKey', key: 'publicKey', ellipsis: true },
  { title: 'CA', dataIndex: 'ca', key: 'ca' },
  {
    title: '当前状态',
    dataIndex: 'status',
    key: 'status',
    customRender: ({ text }: { text: CertItem['status'] }) => {
      const map: Record<CertItem['status'], { color: string; label: string }> = {
        pending: { color: 'orange', label: '待审核' },
        issued: { color: 'green', label: '已签发' },
        rejected: { color: 'red', label: '已拒绝' },
      }
      const m = map[text] || { color: 'default', label: String(text) }

      return {
        children: h(
            resolveComponent('a-tag'),
            { color: m.color },
            { default: () => m.label }
        ),
      }
    },
  },
  { title: '证书说明', dataIndex: 'description', key: 'description', ellipsis: true },
]

// ============ 弹窗表单状态 ============
const modalOpen = ref(false)
const submitLoading = ref(false)

const formRef = ref()
const formModel = reactive({
  publicKey: '',
  description: '',
  csrFileList: [] as any[], // a-upload 受控文件列表
})

// 表单校验规则
const rules = {
  publicKey: [
    { required: true, message: '请输入公钥', trigger: 'blur' },
    { min: 16, message: '公钥长度过短，请检查', trigger: 'blur' },
  ],
  description: [{ required: true, message: '请填写证书说明', trigger: 'blur' }],
  csrFileList: [
    { required: true, type: 'array', min: 1, message: '请上传 CSR 文件', trigger: 'change' },
  ],
}

// ============ 事件函数 ============
function openApply() {
  modalOpen.value = true
}

function resetForm() {
  formModel.publicKey = ''
  formModel.description = ''
  formModel.csrFileList = []
}

// CSR 仅本地选择，不自动上传；在提交时统一打包 FormData
function beforeUpload(_file: File) {
  return false // 阻止 antd 自动上传
}

function onRemove(_file: File) {
  // 交给 a-upload 自己管理，通常不需要额外逻辑
  return true
}

// 拉取证书列表
async function fetchList() {
  loading.value = true
  try {
    const res = await fetch('/api/request/list')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    dataSource.value = Array.isArray(data.items) ? data.items : []
  } catch (e: any) {
    console.error(e)
    message.error('加载证书列表失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  // @ts-ignore
  await formRef.value?.validate().catch(() => { throw new Error('验证失败') })

  // 组装 FormData
  const fd = new FormData()
  fd.append('publicKey', formModel.publicKey.trim())
  fd.append('description', formModel.description.trim())
  // 取第一个 CSR 文件（也可支持多文件）
  const csr = formModel.csrFileList[0]?.originFileObj
  if (!csr) {
    message.warning('请先选择 CSR 文件')
    return
  }
  fd.append('csr', csr) // 字段名 csr，后端按此取文件

  submitLoading.value = true
  try {
    const res = await fetch('/api/cert/request', {
      method: 'POST',
      body: fd, // multipart/form-data 由浏览器自动设置 boundary
      // 注意：不要手动设置 Content-Type
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || '证书申请已提交')
    modalOpen.value = false
    resetForm()
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error('提交申请失败，请稍后重试')
  } finally {
    submitLoading.value = false
  }
}

onMounted(fetchList)
</script>


<template>
  <a-card title="证书申请">
    <!-- 独立按钮区域：位于标题下、列表上 -->
    <div class="apply-actions">
      <a-button type="primary" @click="openApply">
        <i class="fas fa-file-signature" style="margin-right:6px;"></i>
        申请证书
      </a-button>
    </div>

    <a-divider style="margin: 12px 0;" />

    <!-- 证书信息列表 -->
    <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        row-key="id"
        size="middle"
        bordered
        :pagination="{ pageSize: 10 }"
    />
  </a-card>

  <!-- 申请证书表单弹窗 -->
  <a-modal
      v-model:open="modalOpen"
      title="申请证书"
      :confirmLoading="submitLoading"
      ok-text="提交申请"
      cancel-text="取消"
      @ok="handleSubmit"
      @cancel="() => { resetForm() }"
      destroyOnClose
  >
    <a-form
        ref="formRef"
        :model="formModel"
        :rules="rules"
        layout="vertical"
    >
      <a-form-item label="公钥（PEM 或 HEX）" name="publicKey">
        <a-input
            v-model:value="formModel.publicKey"
            placeholder="请输入公钥（例如 04AF... 或 -----BEGIN PUBLIC KEY-----）"
            allow-clear
        />
      </a-form-item>

      <a-form-item label="证书说明" name="description">
        <a-textarea
            v-model:value="formModel.description"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            placeholder="填写申请用途、联系人等信息"
            allow-clear
        />
      </a-form-item>

      <a-form-item label="CSR 文件（.csr/.pem）" name="csrFileList">
        <a-upload
            v-model:file-list="formModel.csrFileList"
            :before-upload="beforeUpload"
            :max-count="1"
            accept=".csr,.pem"
            @remove="onRemove"
            list-type="text"
        >
          <a-button>
            <i class="fas fa-upload" style="margin-right:6px;"></i>
            选择文件
          </a-button>
        </a-upload>
        <div style="color:#999; margin-top:6px;">仅支持单文件，提交时与表单一并上传</div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<style scoped>
.apply-actions {
  display: flex;
  justify-content: flex-start;
  margin: 4px 0 8px;
}
</style>

<style scoped>
.apply-actions {
  display: flex;
  justify-content: flex-start;
  margin: 4px 0 8px;
}
</style>