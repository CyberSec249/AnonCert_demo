<script setup lang="ts">
import { ref, onMounted, h, resolveComponent, reactive } from 'vue'
import { message } from 'ant-design-vue'

type CertItem = {
  id: number
  publicKey: string
  ca_id: number
  caPublicKey: string
  status: 'pending' | 'issued' | 'rejected' | 'waitCRT' | 'acceptCRT'
  description: string
}

const loading = ref(false)
const dataSource = ref<CertItem[]>([])

const columns = [
  { title: '主体公钥',
    dataIndex: 'publicKey',
    key: 'publicKey',
    ellipsis: true,
    align: 'center',
    customRender: ({ text }: { text: string }) => {
      return {
        children: h(
            resolveComponent('a-typography-text'),
            {
              // 悬浮提示完整公钥
              ellipsis: { tooltip: text },
              // 复制按钮：复制的是完整公钥，而不是截断后的
              copyable: { text },
            },
            { default: () => formatKey(text) } // 单元格里显示截断版
        )
      }
    },
  },
  { title: 'CA公钥',
    dataIndex: 'caPublicKey',
    key: 'caPublicKey',
    ellipsis: true,
    align: 'center',
    customRender: ({ record }: { record: any }) => {
      const full = record?.caPublicKey ?? record?.CAPublicKey ?? ''
      return {
        children: h(
            resolveComponent('a-typography-text'),
            {
              // 悬浮提示完整公钥
              ellipsis: { tooltip: full },
              // 复制按钮：复制的是完整公钥，而不是截断后的
              copyable: { text: full },
            },
            { default: () => formatKey(full) } // 单元格里显示截断版
        )
      }
    },
  },
  {
    title: '当前状态',
    align: 'center',
    dataIndex: 'status',
    key: 'status',
    customRender: ({ text }: { text: CertItem['status'] }) => {
      const map: Record<CertItem['status'], { color?: string; label: string }> = {
        pending:  { color: 'orange', label: '待审核' },
        issued:   { color: 'green',  label: '已签发' },
        rejected: { color: 'red',    label: '已拒绝' },
        waitCRT : {color: 'yellow', label: '已审核，待申请CRT'},
        acceptCRT : {color: 'blue', label: 'CRT已验证，待签发'}
      }
      const m = map[text] || { label: String(text) }
      return { children: h(resolveComponent('a-tag'), { color: m.color }, { default: () => m.label }) }
    },
  },
  { title: '证书说明', dataIndex: 'description', key: 'description', ellipsis: true,align: 'center' },
  // 操作按钮列
  {
    title: '操作',
    key: 'actions',
    align: 'center',
    width: 180,
    customRender: ({ record }: { record: CertItem }) => {
      return h('div', { style: 'display:flex; gap:8px;' }, [
        // 查看详情按钮
        h(resolveComponent('a-button'),
            {
              size: 'small',
              type: 'default',
              onClick: () => handleViewDetail(record)
            },
            { default: () => '查看详情' }
        ),

        // CRT 请求按钮（仅状态为 waitCRT 时可用）
        h(resolveComponent('a-button'),
            {
              size: 'small',
              type: 'primary',
              disabled: record.status !== 'waitCRT',
              loading: crtLoadingId.value === record.id,
              onClick: () => handleRequestCRT(record)
            },
            { default: () => 'CRT请求' }
        )
      ])
    }
  }
]

// —— 表单状态
const modalOpen = ref(false)
const submitLoading = ref(false)
const formRef = ref()

const formModel = reactive({
  publicKey: '',
  subjectInfo: '',   // ← 多行输入的“键=值”或 DN 字符串
  caPublicKey: '',
  description: '',
})

// 校验
const rules = {
  publicKey:   [{ required: true, message: '请输入公钥', trigger: 'blur' }, { min: 16, message: '公钥长度过短', trigger: 'blur' }],
  subjectInfo: [{ required: true, message: '请填写主体信息', trigger: 'blur' }],
  caPublicKey: [{ required: true, message: '请输入CA公钥', trigger: 'blur' }, { min: 16, message: 'CA公钥长度过短', trigger: 'blur' }],
  description: [{ required: true, message: '请填写证书说明', trigger: 'blur' }],
}

function openApply() { modalOpen.value = true }
function resetForm() {
  formModel.publicKey = ''
  formModel.subjectInfo = ''
  formModel.caPublicKey = ''
  formModel.description = ''
}

// —— 把“多行键=值”或已成形的 DN 规范化为 RFC4514 DN
function normalizeSubject(input: string): string {
  const text = input.trim()
  // 若本身像 DN（包含逗号分隔的 =），直接简单归一化空格
  if (/,/.test(text) && /=/.test(text)) {
    return text.split(',').map(s => s.trim()).filter(Boolean).join(', ')
  }
  // 否则按多行 key=value 解析
  const kvs: Record<string, string> = {}
  text.split(/\r?\n/).forEach(line => {
    const m = line.match(/^\s*([A-Za-z]+)\s*=\s*(.+)\s*$/)
    if (m) kvs[m[1]] = m[2]
  })
  // 常见字段的输出顺序（缺哪个就跳过）
  const order = ['C','ST','L','O','OU','CN','emailAddress','UID','SN','GN']
  const parts: string[] = []
  for (const k of order) if (k in kvs && kvs[k]) parts.push(`${k}=${kvs[k]}`)
  // 把不在 order 的其余键也带上
  Object.keys(kvs).forEach(k => { if (!order.includes(k)) parts.push(`${k}=${kvs[k]}`) })
  return parts.join(', ')
}

// —— 列表加载
async function fetchList() {
  loading.value = true
  try {
    const res = await fetch('/api/request/list')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    dataSource.value = Array.isArray(data.items) ? data.items : []
  } catch (e) {
    console.error(e)
    message.error('加载证书列表失败')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  // @ts-ignore
  await formRef.value?.validate().catch(() => { throw new Error('验证失败') })

  // 规范化主体信息（DN）
  const subjectDN = normalizeSubject(formModel.subjectInfo)
  if (!subjectDN) { message.warning('主体信息解析为空，请检查'); return }

  const fd = new FormData()
  fd.append('publicKey',   formModel.publicKey.trim())
  fd.append('subjectInfo', subjectDN)              // ← 传规范化后的 DN
  fd.append('caPublicKey', formModel.caPublicKey.trim())
  fd.append('description', formModel.description.trim())

  submitLoading.value = true
  try {
    const res = await fetch('/api/request/cert', { method: 'POST', body: fd })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || '证书申请已提交')
    modalOpen.value = false
    resetForm()
    await fetchList()
  } catch (e) {
    console.error(e)
    message.error('提交申请失败，请稍后重试')
  } finally {
    submitLoading.value = false
  }
}

// 行级加载：哪个 id 的“CRT请求”正在提交
const crtLoadingId = ref<number | null>(null)

// —— CRT 弹窗状态
const crtModalOpen = ref(false)
const crtSubmitting = ref(false)
const crtFormRef = ref()

// 这里都用字符串承载大整数，避免 JS Number 溢出
const crtModel = reactive({
  id: 0 as number,
  moduli: [] as string[],     // 后端返回的模数（十进制或 0x.. 字符串）
  remainders: [] as string[],   // 用户输入的余数（与 moduli 一一对应）
  x: '' as string,     // 用户输入的唯一解 X
})

// 简单大整数字符串校验（支持 0x 前缀）
function isBigIntStr(s: string) {
  if (!s) return false
  if (s.startsWith('0x') || s.startsWith('0X')) {
    return /^[0-9a-fA-F]+$/.test(s.slice(2))
  }
  return /^[0-9]+$/.test(s)
}

// a < b 的比较（均为大整数字符串）
function ltBigIntStr(a: string, b: string) {
  const A = BigInt(a.startsWith('0x') ? a : a) // BigInt 支持十进制/0x
  const B = BigInt(b.startsWith('0x') ? b : b)
  return A < B
}

// —— 点击“CRT请求”按钮
async function handleRequestCRT(record: CertItem) {
  if (record.status !== 'waitCRT') return
  if (crtLoadingId.value !== null) return

  crtLoadingId.value = record.id
  try {
    const res = await fetch('/api/request/crt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: record.id })
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()

    // 填充模型并打开弹窗
    crtModel.id = data.id ?? record.id
    crtModel.moduli = Array.isArray(data.moduli) ? data.moduli : []
    crtModel.remainders = Array.isArray(data.remainders) ? data.remainders : [] // 初始化同长度空数组
    crtModel.x = String(data.x || '')
    crtModalOpen.value = true
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '获取 CRT 模数失败')
  } finally {
    crtLoadingId.value = null
  }
}

// —— 提交 CRT（把用户输入的余数和唯一解发给后端）
async function submitCRT() {
  // 基础校验
  if (!crtModel.moduli.length) {
    message.error('没有模数，无法提交'); return
  }
  for (let i = 0; i < crtModel.moduli.length; i++) {
    const n = (crtModel.moduli[i] || '').trim()
    const r = (crtModel.remainders[i] || '').trim()
    if (!isBigIntStr(n)) { message.error(`n${i+1} 不是有效大整数`); return }
    if (!isBigIntStr(r)) { message.error(`r${i+1} 不是有效大整数`); return }
    // r < n
    try {
      if (!ltBigIntStr(r, n)) { message.error(`r${i+1} 必须 < n${i+1}`); return }
    } catch {
      message.error(`r${i+1}/n${i+1} 比较失败`); return
    }
  }
  if (!isBigIntStr(crtModel.x.trim())) {
    message.error('唯一解 X 不是有效大整数'); return
  }

  crtSubmitting.value = true
  try {
    const payload = {
      id: crtModel.id,
      moduli: crtModel.moduli.map(s => s.trim()),
      remainders: crtModel.remainders.map(s => s.trim()),
      x: crtModel.x,
    }
    const res = await fetch('/api/request/crt/submit', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || 'CRT 数据已提交')
    crtModalOpen.value = false
    await fetchList() // 刷新表格
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '提交 CRT 数据失败')
  } finally {
    crtSubmitting.value = false
  }
}

// —— 详情弹窗状态
const detailModalOpen = ref(false)
const detailLoading = ref(false)
const detailModel = reactive({
  id: 0,
  publicKey: '',
  subjectInfo: '',
  caPublicKey: '',
  description: '',
  status: '',
  ca_id: ''
})
const statusMap: Record<string, { color?: string; label: string }> = {
  pending:  { color: 'orange', label: '待审核' },
  issued:   { color: 'green',  label: '已签发' },
  rejected: { color: 'red',    label: '已拒绝' },
  waitCRT : {color: 'yellow', label: '已审核，待申请CRT'}
}

async function handleViewDetail(record: CertItem) {
  detailLoading.value = true
  detailModalOpen.value = true
  try {
    const res = await fetch(`/api/request/detail?id=${record.id}`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    Object.assign(detailModel, data) // 将返回的对象直接覆盖到 detailModel
  } catch (e) {
    console.error(e)
    message.error('加载详情失败')
    detailModalOpen.value = false
  } finally {
    detailLoading.value = false
  }
}


// —— 查询 CRT 参数：输入公钥弹窗 & 结果弹窗
const queryCRTModalOpen = ref(false)
const queryCRTSubmitting = ref(false)
const queryCRTResultModalOpen = ref(false)

// 输入框模型
const queryCRTForm = reactive({
  publicKey: ''
})

// 查询结果（只读展示）
const crtQueryResult = reactive({
  id: 0 as number,
  moduli: [] as string[],
  remainders: [] as string[],
  x: '' as string
})

function openQueryCRT() {
  queryCRTForm.publicKey = ''
  queryCRTModalOpen.value = true
}

async function submitQueryCRT() {
  const pk = queryCRTForm.publicKey.trim()
  if (!pk) {
    message.warning('请输入公钥');
    return
  }
  queryCRTSubmitting.value = true
  try {
    const res = await fetch('/api/request/crt/query', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ publicKey: pk })
    })
    if(!res.ok) {
      if(res.status === 404) {
        message.warning( '未查询到相关结果')
        return
      }
      throw new Error(`HTTP ${res.status}`)
    }

    const data = await res.json().catch(() => ({}))
    if (data?.found === false) {
      message.warning(data?.msg || '未查询到相关结果')
      return
    }

    const payload = data?.data ?? data
    // 绑定结果（容错处理）
    crtQueryResult.id = payload?.id ?? 0
    crtQueryResult.moduli = Array.isArray(payload?.moduli) ? payload.moduli : []
    crtQueryResult.remainders = Array.isArray(payload?.remainders) ? payload.remainders : []
    crtQueryResult.x = payload?.x ?? ''

    // 关闭输入框，打开结果框
    queryCRTModalOpen.value = false
    queryCRTResultModalOpen.value = true
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '查询 CRT 参数失败')
  } finally {
    queryCRTSubmitting.value = false
  }
}

onMounted(fetchList)

function formatKey(key: string) {
  if (!key) return ''
  const s = String(key)
  // 可按需调整保留长度
  return s.length > 30 ? `${s.slice(0, 12)}...${s.slice(-10)}` : s
}

</script>

<template>
  <a-card title="证书申请列表">
    <div class="apply-actions">
      <a-button type="primary" @click="openApply">
        <i class="fas fa-file-signature" style="margin-right:6px;"></i>申请证书
      </a-button>

      <!-- 查询 CRT 参数按钮 -->
      <a-button type="primary"
                @click="openQueryCRT"
                style="margin-left: 20px; background-color: #2e8b57; border-color: #2e8b57;">
        <i class="fas fa-search" style="margin-right:6px;"></i>查询CRT参数
      </a-button>
    </div>

    <a-divider style="margin: 12px 0;" />

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
    <a-form ref="formRef" :model="formModel" :rules="rules" layout="vertical">
      <a-form-item label="公钥（PEM 或 HEX）" name="publicKey">
        <a-textarea v-model:value="formModel.publicKey"
                    :auto-size="{ minRows: 2, maxRows: 4 }"
                    placeholder="04AF... 或 -----BEGIN PUBLIC KEY-----"
                    allow-clear />
      </a-form-item>

      <!-- 统一输入框：主体信息（支持多行键=值 / 或直接 DN） -->
      <a-form-item label="主体信息（多行键=值 或 DN）" name="subjectInfo">
        <a-textarea
            v-model:value="formModel.subjectInfo"
            :auto-size="{ minRows: 3, maxRows: 8 }"
            placeholder="示例：&#10;C=CN&#10;ST=Beijing&#10;L=Haidian&#10;O=Huawei&#10;OU=DPKI&#10;CN=example.com&#10;emailAddress=admin@example.com"
            allow-clear
        />
        <div style="color:#999; margin-top:6px;">
          也可直接粘贴 DN：C=CN, ST=Beijing, L=Haidian, O=Huawei, OU=DPKI, CN=example.com, emailAddress=admin@example.com
        </div>
      </a-form-item>

      <a-form-item label="CA 公钥（PEM 或 HEX）" name="caPublicKey">
        <a-textarea
            v-model:value="formModel.caPublicKey"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            placeholder="04AF... 或 -----BEGIN PUBLIC KEY-----"
            allow-clear
        />
      </a-form-item>

      <a-form-item label="证书说明" name="description">
        <a-textarea
            v-model:value="formModel.description"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            placeholder="用途、联系人等信息"
            allow-clear
        />
      </a-form-item>
    </a-form>
  </a-modal>


  <a-modal
      v-model:open="detailModalOpen"
      title="申请信息"
      :footer="null"
      :confirmLoading="detailLoading"
      destroyOnClose
  >
    <a-form layout="vertical">
      <a-form-item label="ID" hidden="hidden">
        <a-input v-model:value="detailModel.id" readonly />
      </a-form-item>

      <a-form-item label="公钥">
        <a-textarea v-model:value="detailModel.publicKey" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

      <a-form-item label="主体信息">
        <a-textarea v-model:value="detailModel.subjectInfo" :auto-size="{ minRows: 2, maxRows: 6 }" readonly />
      </a-form-item>

      <a-form-item label="CA">
        <a-textarea v-model:value="detailModel.caPublicKey" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

      <a-form-item label="证书说明">
        <a-textarea v-model:value="detailModel.description" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

      <a-form-item label="当前申请进度">
        <a-input :value="statusMap[detailModel.status]?.label || detailModel.status" readonly />
      </a-form-item>

    </a-form>
  </a-modal>


  <!-- CRT 请求弹窗：显示模数（只读），输入余数与唯一解，提交后端 -->
  <a-modal
      v-model:open="crtModalOpen"
      title="CRT 请求"
      :confirmLoading="crtSubmitting"
      ok-text="提交"
      cancel-text="取消"
      @ok="submitCRT"
      destroyOnClose
      width="720px"
  >
    <a-form ref="crtFormRef" layout="vertical">
      <!-- 模数只读列表 -->
      <a-form-item label="模数 nᵢ（只读）">
        <div style="max-height:220px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
          <div v-for="(n, idx) in crtModel.moduli" :key="'n-'+idx" style="margin-bottom:10px;">
            <div><b>CA{{ idx+1 }} -> n{{ idx+1 }}：</b></div>
            <a-textarea
                :value="n"
                readonly
                :auto-size="{ minRows: 1, maxRows: 3 }"
            />
          </div>
        </div>
      </a-form-item>

      <!-- 余数输入列表，与模数一一对应 -->
      <a-form-item label="余数 rᵢ（与上方 nᵢ 对应）">
        <div style="max-height:220px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
          <div v-for="(n, idx) in crtModel.remainders" :key="'r-'+idx" style="margin-bottom:10px;">
            <div><b>r{{ idx+1 }}（大整数字符串，需满足 r{{idx+1}} &lt; n{{idx+1}}）：</b></div>
            <a-input
                v-model:value="crtModel.remainders[idx]"
                placeholder="请输入 rᵢ（十进制或 0x..）"
                allow-clear
            />
          </div>
        </div>
      </a-form-item>

      <!-- 唯一解 X -->
      <a-form-item label="唯一解 X（大整数字符串）">
        <a-input
            v-model:value="crtModel.x"
            placeholder="点击<CRT计算>唯一解 X"
            readonly
        />
      </a-form-item>
    </a-form>
  </a-modal>


  <!-- （一）输入公钥的弹窗 -->
  <a-modal
      v-model:open="queryCRTModalOpen"
      title="查询 CRT 参数"
      :confirmLoading="queryCRTSubmitting"
      ok-text="查询"
      cancel-text="取消"
      @ok="submitQueryCRT"
      destroyOnClose
  >
    <a-form layout="vertical">
      <a-form-item label="公钥（PEM 或 HEX）">
        <a-textarea
            v-model:value="queryCRTForm.publicKey"
            :auto-size="{ minRows: 2, maxRows: 4 }"
            placeholder="请输入要查询的公钥"
            allow-clear
        />
        <!-- MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAELASPyL5lNtfVvTIJZIZqIim9Pkb+h8DShVdYXW4xO6+vZtUaoJDlIKKNFoBsnG9JDRoFqHqM+u0V6CK8dwTn5w== -->
      </a-form-item>
      <div style="color:#999;">说明：将根据公钥查询对应的 CRT 参数。</div>
    </a-form>
  </a-modal>

  <!-- （二）结果展示弹窗（只读） -->
  <a-modal
      v-model:open="queryCRTResultModalOpen"
      title="CRT 参数"
      :footer="null"
      destroyOnClose
      width="720px"
  >
    <a-form layout="vertical">

      <a-form-item label="模数 nᵢ（只读）">
        <div style="max-height:240px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
          <div v-for="(n, idx) in crtQueryResult.moduli" :key="'q-n-'+idx" style="margin-bottom:10px;">
            <div><b>n{{ idx+1 }}：</b></div>
            <a-textarea
                :value="n"
                readonly
                :auto-size="{ minRows: 1, maxRows: 3 }"
            />
          </div>
          <div v-if="!crtQueryResult.moduli.length" style="color:#999;">无模数数据</div>
        </div>
      </a-form-item>

      <a-form-item label="余数 nᵢ（只读）">
        <div style="max-height:240px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
          <div v-for="(n, idx) in crtQueryResult.remainders" :key="'q-n-'+idx" style="margin-bottom:10px;">
            <div><b>n{{ idx+1 }}：</b></div>
            <a-textarea
                :value="n"
                readonly
                :auto-size="{ minRows: 1, maxRows: 3 }"
            />
          </div>
          <div v-if="!crtQueryResult.remainders.length" style="color:#999;">无余数数据</div>
        </div>
      </a-form-item>

      <a-form-item label="唯一解 X（只读）">
        <a-input :value="crtQueryResult.x" readonly />
      </a-form-item>
    </a-form>
  </a-modal>



</template>
