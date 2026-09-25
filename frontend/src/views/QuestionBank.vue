<template>
  <div class="page-card">
    <div class="page-header">
      <h2>题库管理</h2>
      <div>
        <el-button type="success" @click="onBatchImport">批量导入</el-button>
        <el-button type="primary" @click="openCreate">新增题目</el-button>
      </div>
    </div>

    <div class="toolbar">
      <el-select v-model="query.type" placeholder="题型" clearable style="width: 150px">
        <el-option v-for="(label, key) in typeLabels" :key="key" :label="label" :value="key" />
      </el-select>
      <el-select v-model="query.difficulty" placeholder="难度" clearable style="width: 130px">
        <el-option v-for="(label, key) in difficultyLabels" :key="key" :label="label" :value="key" />
      </el-select>
      <el-input v-model="query.knowledge_point" placeholder="知识点" clearable style="width: 180px" />
      <el-input v-model="query.keyword" placeholder="题干关键词" clearable style="width: 180px" />
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="题型" width="100">
        <template #default="{ row }">{{ typeLabels[row.type as keyof typeof typeLabels] || row.type }}</template>
      </el-table-column>
      <el-table-column prop="content" label="题干" min-width="220" show-overflow-tooltip />
      <el-table-column label="难度" width="90">
        <template #default="{ row }">
          <el-tag :type="difficultyTag[row.difficulty as keyof typeof difficultyTag]">{{ difficultyLabels[row.difficulty as keyof typeof difficultyLabels] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="knowledge_point" label="知识点" width="130" />
      <el-table-column label="分值" width="110">
        <template #default="{ row }">
          <div>{{ row.score }}</div>
          <div v-if="row.scoring_points && row.scoring_points.length" class="cell-sub">
            {{ row.scoring_points.length }} 个评分点
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="onDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      class="pager"
      v-model:current-page="query.page"
      v-model:page-size="query.page_size"
      :total="total"
      layout="total, prev, pager, next"
      @current-change="load"
    />

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑题目' : '新增题目'" width="700px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="题型">
          <el-select v-model="form.type" :disabled="!!editingId" @change="onTypeChange">
            <el-option v-for="(label, key) in typeLabels" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="题干">
          <el-input v-model="form.content" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="选项" v-if="isChoice">
          <div style="width: 100%">
            <div v-for="(opt, idx) in form.options" :key="idx" class="option-row">
              <el-input v-model="opt.key" placeholder="key" style="width: 80px" />
              <el-input v-model="opt.text" placeholder="选项内容" style="flex: 1" />
              <el-button link type="danger" @click="form.options.splice(idx, 1)">删除</el-button>
            </div>
            <el-button size="small" @click="addOption">添加选项</el-button>
          </div>
        </el-form-item>
        <el-form-item label="答案">
          <el-select v-if="form.type === 'single'" v-model="form.answerSingle" placeholder="选择正确答案">
            <el-option v-for="opt in form.options" :key="opt.key" :label="`${opt.key}. ${opt.text}`" :value="opt.key" />
          </el-select>
          <el-select v-else-if="form.type === 'multiple'" v-model="form.answerMultiple" multiple placeholder="选择正确答案">
            <el-option v-for="opt in form.options" :key="opt.key" :label="`${opt.key}. ${opt.text}`" :value="opt.key" />
          </el-select>
          <el-radio-group v-else-if="form.type === 'true_false'" v-model="form.answerSingle">
            <el-radio value="T">正确</el-radio>
            <el-radio value="F">错误</el-radio>
          </el-radio-group>
          <div v-else-if="form.type === 'fill_blank'" style="width: 100%">
            <div v-for="(blank, idx) in form.answerBlanks" :key="idx" class="option-row">
              <el-input v-model="form.answerBlanks[idx]" placeholder="第 {{ idx + 1 }} 空参考答案" />
              <el-button link type="danger" @click="form.answerBlanks.splice(idx, 1)">删除</el-button>
            </div>
            <el-button size="small" @click="form.answerBlanks.push('')">添加空</el-button>
          </div>
          <el-input v-else v-model="form.answerText" type="textarea" placeholder="参考答案" />
        </el-form-item>
        <el-form-item label="解析">
          <el-input v-model="form.analysis" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="难度">
          <el-select v-model="form.difficulty">
            <el-option v-for="(label, key) in difficultyLabels" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="知识点">
          <el-input v-model="form.knowledge_point" />
        </el-form-item>
        <el-form-item v-if="isSubjective" label="评分点">
          <div style="width: 100%">
            <div v-if="!form.scoring_points.length" class="hint-text">
              未设置评分点，将按整题给分（兼容旧题）。
            </div>
            <div v-for="(point, idx) in form.scoring_points" :key="idx" class="option-row">
              <el-input v-model="point.name" placeholder="要点名称，如：第一空/踩分点" style="flex: 1" />
              <el-input-number v-model="point.score" :min="0.5" :step="0.5" />
              <el-button link type="danger" @click="form.scoring_points.splice(idx, 1)">删除</el-button>
            </div>
            <el-button size="small" @click="form.scoring_points.push({ name: '', score: 1 })">添加评分点</el-button>
            <span v-if="form.scoring_points.length" class="hint-text">
              评分点合计：{{ pointsTotal }} 分
            </span>
          </div>
        </el-form-item>
        <el-form-item label="分值">
          <el-input-number v-if="isSubjective && form.scoring_points.length" :model-value="pointsTotal" disabled />
          <el-input-number v-else v-model="form.score" :min="0.5" :step="0.5" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="onSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { questionApi } from '../api'
import type { Question, QuestionType } from '../types'

const typeLabels: Record<string, string> = {
  single: '单选题',
  multiple: '多选题',
  true_false: '判断题',
  fill_blank: '填空题',
  short_answer: '简答题'
}
const difficultyLabels: Record<string, string> = { easy: '简单', medium: '中等', hard: '困难' }
const difficultyTag: Record<string, string> = { easy: 'success', medium: 'warning', hard: 'danger' }

const loading = ref(false)
const saving = ref(false)
const rows = ref<Question[]>([])
const total = ref(0)
const dialogVisible = ref(false)
const editingId = ref<number | null>(null)
const query = reactive({ page: 1, page_size: 10, type: '', difficulty: '', knowledge_point: '', keyword: '' })

const form = reactive({
  type: 'single' as QuestionType,
  content: '',
  options: [] as { key: string; text: string }[],
  answerSingle: '',
  answerMultiple: [] as string[],
  answerBlanks: [''],
  answerText: '',
  analysis: '',
  difficulty: 'easy',
  knowledge_point: '',
  score: 1,
  scoring_points: [] as { name: string; score: number }[]
})

const isChoice = () => ['single', 'multiple', 'true_false'].includes(form.type)
const isSubjective = computed(() => ['fill_blank', 'short_answer'].includes(form.type))
const pointsTotal = computed(() =>
  Math.round(form.scoring_points.reduce((sum, p) => sum + (Number(p.score) || 0), 0) * 100) / 100
)

function addOption() {
  const idx = form.options.length
  form.options.push({ key: String.fromCharCode(65 + idx), text: '' })
}

function onTypeChange() {
  form.options = form.type === 'true_false' ? [] : form.options
}

function buildPayload() {
  let answer: unknown
  if (form.type === 'single') answer = form.answerSingle
  else if (form.type === 'multiple') answer = form.answerMultiple
  else if (form.type === 'true_false') answer = form.answerSingle
  else if (form.type === 'fill_blank') answer = form.answerBlanks.filter((b) => b.trim() !== '')
  else answer = form.answerText
  const score = isSubjective.value && form.scoring_points.length ? pointsTotal.value : form.score
  const scoringPoints =
    isSubjective.value && form.scoring_points.length
      ? form.scoring_points.map((p) => ({ name: p.name.trim(), score: p.score }))
      : []
  return {
    type: form.type,
    content: form.content,
    options: isChoice() ? form.options : [],
    answer,
    analysis: form.analysis,
    difficulty: form.difficulty as 'easy' | 'medium' | 'hard',
    knowledge_point: form.knowledge_point,
    score,
    scoring_points: scoringPoints
  }
}

function validateForm() {
  if (!isSubjective.value || !form.scoring_points.length) return true
  const names = new Set<string>()
  for (const p of form.scoring_points) {
    if (!p.name.trim()) {
      ElMessage.error('评分点名称不能为空')
      return false
    }
    if (names.has(p.name.trim())) {
      ElMessage.error(`评分点名称重复：${p.name.trim()}`)
      return false
    }
    names.add(p.name.trim())
    if (!(p.score > 0)) {
      ElMessage.error(`评分点「${p.name}」分值必须大于 0`)
      return false
    }
  }
  if (pointsTotal.value <= 0) {
    ElMessage.error('评分点分值合计必须大于 0')
    return false
  }
  return true
}

function resetForm() {
  form.type = 'single'
  form.content = ''
  form.options = []
  form.answerSingle = ''
  form.answerMultiple = []
  form.answerBlanks = ['']
  form.answerText = ''
  form.analysis = ''
  form.difficulty = 'easy'
  form.knowledge_point = ''
  form.score = 1
  form.scoring_points = []
}

function openCreate() {
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(row: Question) {
  editingId.value = row.id
  form.type = row.type
  form.content = row.content
  form.options = (row.options || []).map((o) => ({ key: o.key, text: o.text }))
  if (row.type === 'single' || row.type === 'true_false') form.answerSingle = (row.answer as string) || ''
  else if (row.type === 'multiple') form.answerMultiple = (row.answer as string[]) || []
  else if (row.type === 'fill_blank') form.answerBlanks = (row.answer as string[]) || ['']
  else form.answerText = (row.answer as string) || ''
  form.analysis = row.analysis || ''
  form.difficulty = row.difficulty
  form.knowledge_point = row.knowledge_point
  form.scoring_points = (row.scoring_points || []).map((p) => ({ name: p.name, score: p.score }))
  form.score = row.score
  dialogVisible.value = true
}

async function onSave() {
  if (!validateForm()) return
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await questionApi.update(editingId.value, payload)
    } else {
      await questionApi.create(payload)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    load()
  } finally {
    saving.value = false
  }
}

async function onDelete(row: Question) {
  await ElMessageBox.confirm('确认删除该题目？', '提示', { type: 'warning' })
  await questionApi.remove(row.id)
  ElMessage.success('删除成功')
  load()
}

function onBatchImport() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json,.xlsx'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    const res = await questionApi.batchImport(file)
    ElMessage.success(`导入完成：成功 ${res.imported} 条，失败 ${res.failed} 条`)
    load()
  }
  input.click()
}

async function load() {
  loading.value = true
  try {
    const res = await questionApi.list({ ...query })
    rows.value = res.items
    total.value = res.total
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.option-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}
.pager {
  margin-top: 16px;
  justify-content: flex-end;
}
.hint-text {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-left: 8px;
}
.cell-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
