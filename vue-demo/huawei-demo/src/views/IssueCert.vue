<script setup lang="ts">
import { ref, reactive, onMounted, h, resolveComponent } from 'vue'
import { message } from 'ant-design-vue'

type CRTParams = {
  moduli: string[]
  remainders: string[]
  x: string
}

type IssueItem = {
  request_id: number
  publicKey: string
  subjectInfo: string
  crt: CRTParams
  subject_status: boolean | number | string
  crt_status: boolean | number | string
  if_reject: boolean | number | string
}

const centerCell = () => ({style: {textAlign: 'center'}})

const loading = ref(false)
const dataSource = ref<IssueItem[]>([])

// 行级 loading（避免整表锁死）
const rowLoading = reactive<Record<number, boolean>>({})

// —— 弹窗：查看主体信息 & 查看 CRT 参数
const subjectViewOpen = ref(false)
const crtViewOpen = ref(false)
const subjectFetchLoading = ref(false)
const crtFetchLoading = ref(false)

// 当前在弹窗中查看/审批/验证的记录（主体）
const currentSubject = reactive<{ id: number; publicKey: string; subjectInfo: string }>({
  id: 0, publicKey: '', subjectInfo: ''
})
const subjectView = ref('')
const subjectApproveLoading = ref(false)

// 当前在弹窗中查看/验证的记录（CRT）
const currentCRT = reactive<{ id: number; publicKey: string; crt: CRTParams }>({
  id: 0, publicKey: '', crt: { moduli: [], remainders: [], x: '' }
})
const crtView = reactive<CRTParams>({ moduli: [], remainders: [], x: '' })
const crtApproveLoading = ref(false)

// —— 核验/验证结果（保留一个通用弹窗，如后续需要提示更详细信息可复用）
const verifyOpen = ref(false)
const verifyTitle = ref('')
const verifyResult = reactive<{ ok: boolean; detail: string }>({ ok: false, detail: '' })

// 布尔归一化（后端可能返回 0/1 或 "0"/"1"）
function toBool(v: unknown): boolean {
  if (typeof v === 'boolean') return v
  if (typeof v === 'number')  return v === 1
  if (typeof v === 'string')  return v === '1' || v.toLowerCase() === 'true'
  return false
}

// 列定义
const columns = [
  {
    title: '公钥',
    dataIndex: 'publicKey',
    key: 'publicKey',
    align: 'center',
    width: 240,
    ellipsis: true,
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
    customRender: ({ text }: { text: string }) =>
        h(resolveComponent('a-tooltip'), { title: text },
            { default: () => h('span', text) }
        )
  },
  {
    title: '主体信息',
    key: 'subjectInfo',
    align: 'center',
    width: 220,
    customCell: centerCell,
    customRender: ({ record }: { record: IssueItem }) => {
      return h('div', { style: 'display:flex; gap:8px; justify-content:center; width:100%;  flex-wrap: wrap;' }, [
        h(resolveComponent('a-button'),
            { size: 'small', onClick: () => handleViewSubject(record) },
            { default: () => '查看主体信息' }
        )
      ])
    }
  },
  {
    title: 'CRT参数',
    key: 'crt',
    width: 220,
    align: 'center',
    customCell: centerCell,
    customRender: ({ record }: { record: IssueItem }) => {
      return h('div', { style: 'display:flex; gap:8px; justify-content:center; width:100%; flex-wrap: wrap;' }, [
        h(resolveComponent('a-button'),
            { size: 'small', onClick: () => handleViewCRT(record) },
            { default: () => '查看CRT参数' }
        )
      ])
    }
  },
  {
    title: '当前审核进度',
    key: 'progress',
    width: 360,
    align: 'center',
    customRender: ({ record }: { record: IssueItem }) => {
      const okSubject = toBool(record.subject_status)
      const okCRT     = toBool(record.crt_status)
      const okReject = toBool(record.if_reject)

      let label = ''
      let color: 'red' | 'green' | 'default' = 'default'
      if (!okSubject && !okCRT && !okReject) {
        label = '主体信息未审核，CRT参数未验证'
        color = 'default'
      } else if (okSubject && !okCRT && !okReject) {
        label = '主体信息已审核，CRT参数未验证'
        color = 'default'
      } else if (okReject) {
        label = '已拒绝签发请求'
        color = 'red'
      } else {
        label = '主体信息已审核，CRT参数已验证'
        color = 'green'
      }

      return {
        children: h(resolveComponent('a-tag'), { color }, { default: () => label })
      }
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    align: 'center',
    customCell: centerCell,
    customRender: ({ record }: { record: IssueItem }) => {
      const okSubject = toBool(record.subject_status)
      const okCRT     = toBool(record.crt_status)
      const okReject  = toBool(record.if_reject)
      const disabled  = !(okSubject && okCRT) || okReject

      return h('div', { style: 'display:flex; gap:8px; justify-content:center; width:100%;' }, [
        h(
            resolveComponent('a-button'),
            {
              type: 'primary',
              disabled,
              loading: !!rowLoading[record.request_id],
              onClick: () => issueAnonymousCert(record),
            },
            { default: () => '签发匿名证书' }
        )
      ])
    }
  }
]

// —— 加载待签发列表
async function fetchList() {
  loading.value = true
  try {
    const res = await fetch('/api/issue/list')   // 替换为你的实际接口
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    const items = Array.isArray(data.items) ? data.items : []
    // 归一化布尔字段，避免 "0"/"1" 的坑
    dataSource.value = items.map((it: IssueItem) => ({
      ...it,
      subject_status: toBool(it.subject_status),
      crt_status:     toBool(it.crt_status),
      if_reject:    toBool(it.if_reject),
    }))
  } catch (e) {
    console.error(e)
    message.error('加载待签发列表失败')
  } finally {
    loading.value = false
  }
}

// —— 查看主体信息（并在弹窗中提供“通过审核”）
async function handleViewSubject(record: IssueItem) {
  subjectFetchLoading.value = true
  subjectViewOpen.value = true
  try {
    const res = await fetch(`/api/issue/subject/detail`,{
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: String(record.request_id) })
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    const payload = data?.data ?? data

    subjectView.value = payload?.subjectInfo ?? ''

    // 回填当前记录（确保后续“通过/拒绝”能携带 id）
    currentSubject.id = record.request_id
    currentSubject.publicKey = record.publicKey
    currentSubject.subjectInfo = subjectView.value

  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '加载主体信息失败')
    // 拉取失败可以关闭弹窗或保留空态
    subjectViewOpen.value = false
  } finally {
    subjectFetchLoading.value = false
  }
}

async function approveSubject() {
  if (!currentSubject.id) return
  subjectApproveLoading.value = true
  try {
    const res = await fetch('/api/issue/cert/check', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        request_id: currentSubject.id,
        subject_status: "1",
        crt_status: null
      })
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || '主体信息核验已通过')
    subjectViewOpen.value = false
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '主体信息核验失败')
  } finally {
    subjectApproveLoading.value = false
  }
}

// —— 查看 CRT 参数（并在弹窗中提供“通过验证”）
async function handleViewCRT(record: IssueItem) {
  crtFetchLoading.value = true

  try {
    const res = await fetch('/api/issue/crt/detail', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ request_id: String(record.request_id) })
    })

    const data = await res.json().catch(() => ({}))
    if (data?.found === false) {
      message.warning(data?.msg || '未查询到相关结果')
      return
    }

    crtViewOpen.value = true  // 先打开弹窗，配合 loading

    const payload = data?.data ?? data
    // 绑定结果（容错）
    const moduli     = Array.isArray(payload?.moduli) ? payload.moduli : []
    const remainders = Array.isArray(payload?.remainders) ? payload.remainders : []
    const x          = payload?.x ?? ''

    crtView.moduli = [...moduli]
    crtView.remainders = [...remainders]
    crtView.x = x

    // 回填当前请求标识（后续“通过/拒绝”用）
    currentCRT.id = record.request_id   // 与上面的 body 保持同一含义
    currentCRT.publicKey = record.publicKey
    currentCRT.crt = { moduli: [...crtView.moduli], remainders: [...crtView.remainders], x: crtView.x }

  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '加载 CRT 参数失败')
    crtViewOpen.value = false
  } finally {
    crtFetchLoading.value = false
  }
}

async function verifyCRT() {
  if (!currentCRT.id) return
  crtApproveLoading.value = true
  try {
    const res = await fetch('/api/issue/crt/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        id: currentCRT.id,
        moduli: currentCRT.crt.moduli.map(String),
        remainders: currentCRT.crt.remainders.map(String),
        x: String(currentCRT.crt.x),
      })
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || 'CRT 参数验证成功')
    crtViewOpen.value = false
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || 'CRT 参数验证失败')
  } finally {
    crtApproveLoading.value = false
  }
}

async function approveCRT() {
  if (!currentCRT.id) return
  crtApproveLoading.value = true

  try {
    const res = await fetch('/api/issue/cert/check', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        request_id: currentCRT.id,
        subject_status: null,
        crt_status: "1"
      })
    })

    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.success(data?.msg || 'CRT 请求已通过')
    crtViewOpen.value = false
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || 'CRT 请求失败')
  } finally {
    crtApproveLoading.value = false
  }
}

// 拒绝 CRT 参数
async function rejectCRT() {
  if (!currentCRT.id) return
  crtApproveLoading.value = true
  try {
    const res = await fetch('/api/issue/cert/reject', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ request_id: currentCRT.id })
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.warning(data?.msg || 'CRT 请求未已拒绝')
    crtViewOpen.value = false
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || 'CRT 请求拒绝失败')
  } finally {
    crtApproveLoading.value = false
  }
}

// 拒绝主体信息
async function rejectSubject() {
  if (!currentSubject.id) return
  subjectApproveLoading.value = true
  try {
    const res = await fetch('/api/issue/cert/reject', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ request_id: currentSubject.id })
    })

    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json().catch(() => ({}))
    message.warning(data?.msg || '主体信息核验请求已拒绝')
    subjectViewOpen.value = false
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || 'CRT 主体信息拒绝失败')
  } finally {
    subjectApproveLoading.value = false
  }
}

async function issueAnonymousCert(record: IssueItem) {
  const id = record.request_id
  if (!id) {
    message.warning('无效的 request_id')
    return
  }

  // 行级 loading
  rowLoading[id] = true
  try {
    const res = await fetch('/api/issue/cert/issuance', {
      method: 'POST', // 如果你后端是其它动词/路径，请改这里
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ request_id: id })
    })

    if (!res.ok) {
      const txt = await res.text().catch(() => '')
      throw new Error(`HTTP ${res.status} ${txt}`)
    }

    // 兼容后端可能没有返回 JSON 的情况
    let data: any = {}
    try { data = await res.json() } catch {}

    message.success(data?.msg || '匿名证书已签发')
    await fetchList()
  } catch (e: any) {
    console.error(e)
    message.error(e?.message || '签发匿名证书失败')
  } finally {
    rowLoading[id] = false
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
  <a-card title="证书签发">
    <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        row-key="request_id"
        bordered
        size="middle"
        :pagination="{ pageSize: 10 }"
    />

    <!-- 查看主体信息 -->
    <a-modal
        v-model:open="subjectViewOpen"
        title="主体信息"
        :footer="null"
        destroyOnClose
    >
      <a-textarea :value="subjectView" :auto-size="{ minRows: 6, maxRows: 16 }" readonly />
      <div style="margin-top:12px; display:flex; justify-content:flex-end; gap:8px;">
        <a-button danger :loading="subjectApproveLoading" @click="rejectSubject">
          拒绝通过
        </a-button>
        <a-button type="primary" :loading="subjectApproveLoading" @click="approveSubject">
          通过审核
        </a-button>
      </div>
    </a-modal>

    <!-- 查看 CRT 参数（弹窗内提供：通过验证） -->
    <a-modal
        v-model:open="crtViewOpen"
        title="CRT 参数"
        :footer="null"
        width="720px"
        destroyOnClose
    >
      <a-form layout="vertical">
        <a-form-item label="模数 nᵢ（只读）">
          <div style="max-height:240px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
            <div v-for="(n, idx) in crtView.moduli" :key="'view-n-'+idx" style="margin-bottom:10px;">
              <div><b>n{{ idx+1 }}：</b></div>
              <a-textarea :value="n" readonly :auto-size="{ minRows: 1, maxRows: 3 }" />
            </div>
            <div v-if="!crtView.moduli.length" style="color:#999;">无模数数据</div>
          </div>
        </a-form-item>

        <a-form-item label="余数 rᵢ（只读）">
          <div style="max-height:240px; overflow:auto; border:1px solid #f0f0f0; border-radius:6px; padding:8px;">
            <div v-for="(r, idx) in crtView.remainders" :key="'view-r-'+idx" style="margin-bottom:10px;">
              <div><b>r{{ idx+1 }}：</b></div>
              <a-textarea :value="r" readonly :auto-size="{ minRows: 1, maxRows: 3 }" />
            </div>
            <div v-if="!crtView.remainders.length" style="color:#999;">无余数数据</div>
          </div>
        </a-form-item>

        <a-form-item label="唯一解 X（只读）">
          <a-input :value="crtView.x" readonly />
        </a-form-item>
      </a-form>

      <div style="margin-top:12px; display:flex; justify-content:flex-end; gap:8px;">
        <a-button type="primary" :loading="crtApproveLoading" @click="verifyCRT">
          验证CRT
        </a-button>
        <a-button danger :loading="crtApproveLoading" @click="rejectCRT">
          拒绝通过
        </a-button>
        <a-button style="color: #207804; border-color: #207804" :loading="crtApproveLoading" @click="approveCRT">
          通过验证
        </a-button>
      </div>
    </a-modal>

    <a-modal v-model:open="verifyOpen" :title="verifyTitle" :footer="null" destroyOnClose>
      <div style="display:flex; align-items:center; gap:10px; margin-bottom:8px;">
        <a-tag :color="verifyResult.ok ? 'green' : 'red'">{{ verifyResult.ok ? '通过' : '未通过' }}</a-tag>
        <span>{{ verifyResult.detail }}</span>
      </div>
    </a-modal>
  </a-card>
</template>
