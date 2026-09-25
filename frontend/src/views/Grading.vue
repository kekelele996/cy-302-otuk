<template>
  <div class="page-card">
    <div class="page-header">
      <h2>主观题批改</h2>
      <el-button @click="$router.push('/exams')">返回</el-button>
    </div>

    <div class="toolbar">
      <span>选择答题记录：</span>
      <el-select v-model="attemptId" placeholder="请选择" style="width: 320px" @change="loadDetail">
        <el-option v-for="a in attempts" :key="a.attempt_id" :label="`#${a.attempt_id} 客观分 ${a.objective_score}`" :value="a.attempt_id" />
      </el-select>
    </div>

    <template v-if="detail">
      <el-table :data="subjectiveQuestions" border>
        <el-table-column prop="content" label="题干" min-width="200" show-overflow-tooltip />
        <el-table-column label="学生答案" min-width="180">
          <template #default="{ row }">
            <div class="answer-text">{{ formatAnswer(row.student_answer) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="参考答案" min-width="180">
          <template #default="{ row }">
            <div class="answer-text">{{ formatAnswer(row.correct_answer) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="评分点录分" min-width="280">
          <template #default="{ row }">
            <div v-if="row.rubric && row.rubric.length">
              <div v-for="(point, idx) in row.rubric" :key="point.name" class="point-row">
                <span class="point-name">{{ point.name }}（满分 {{ point.score }}）</span>
                <el-input-number
                  v-model="pointScores[row.exam_question_id][idx].score"
                  :min="0"
                  :max="point.score"
                  :step="0.5"
                  size="small"
                />
              </div>
              <div class="point-total" :class="{ overflow: pointTotal(row.exam_question_id) > row.max_score }">
                合计 {{ pointTotal(row.exam_question_id) }} / {{ row.max_score }} 分
              </div>
            </div>
            <el-input-number v-else v-model="scores[row.exam_question_id]" :min="0" :max="row.max_score" :step="0.5" />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.graded ? 'success' : 'info'">{{ row.graded ? '已批' : '待批' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-button v-if="subjectiveQuestions.length" type="primary" style="margin-top: 16px" :loading="saving" @click="onSubmit">保存批改</el-button>
      <el-empty v-else description="该试卷没有主观题" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { attemptApi, examApi } from '../api'
import type { AttemptDetail, AttemptSummary, RubricPoint } from '../types'

const route = useRoute()
const examId = Number(route.params.examId)
const attempts = ref<AttemptSummary[]>([])
const attemptId = ref<number | null>(null)
const detail = ref<AttemptDetail | null>(null)
const scores = reactive<Record<number, number>>({})
const pointScores = reactive<Record<number, RubricPoint[]>>({})
const saving = ref(false)

const subjectiveQuestions = computed(() => {
  if (!detail.value) return []
  return detail.value.questions.filter((q) => q.type === 'fill_blank' || q.type === 'short_answer')
})

function formatAnswer(v: unknown) {
  if (v === null || v === undefined) return '-'
  if (Array.isArray(v)) return v.join('；')
  return String(v)
}

function round2(v: number) {
  return Math.round(v * 100) / 100
}

function pointTotal(examQuestionId: number) {
  const points = pointScores[examQuestionId] || []
  return round2(points.reduce((acc, p) => acc + (p.score || 0), 0))
}

async function loadDetail() {
  if (!attemptId.value) return
  detail.value = await attemptApi.detail(attemptId.value)
  detail.value.questions.forEach((q) => {
    if (q.type !== 'fill_blank' && q.type !== 'short_answer') return
    if (q.rubric && q.rubric.length) {
      const awarded = new Map((q.point_scores || []).map((p) => [p.name, p.score]))
      pointScores[q.exam_question_id] = q.rubric.map((p) => ({
        name: p.name,
        score: awarded.get(p.name) ?? 0
      }))
    } else {
      scores[q.exam_question_id] = q.score || 0
    }
  })
}

async function onSubmit() {
  if (!attemptId.value) return
  const items: { exam_question_id: number; score?: number; point_scores?: RubricPoint[] }[] = []
  for (const q of subjectiveQuestions.value) {
    if (q.rubric && q.rubric.length) {
      const points = pointScores[q.exam_question_id] || []
      for (let i = 0; i < q.rubric.length; i++) {
        const max = q.rubric[i].score
        const value = points[i]?.score ?? 0
        if (value < 0 || value > max) {
          ElMessage.error(`「${q.rubric[i].name}」得分不能超过该要点满分 ${max} 分`)
          return
        }
      }
      const total = pointTotal(q.exam_question_id)
      if (total > q.max_score) {
        ElMessage.error(`评分点合计 ${total} 分超过题目满分 ${q.max_score} 分`)
        return
      }
      items.push({
        exam_question_id: q.exam_question_id,
        point_scores: points.map((p) => ({ name: p.name, score: p.score }))
      })
    } else {
      const score = scores[q.exam_question_id] ?? 0
      if (score < 0 || score > q.max_score) {
        ElMessage.error(`得分不能超过题目满分 ${q.max_score} 分`)
        return
      }
      items.push({ exam_question_id: q.exam_question_id, score })
    }
  }
  saving.value = true
  try {
    await attemptApi.grade(attemptId.value, items)
    ElMessage.success('批改成功')
    loadDetail()
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  attempts.value = await examApi.attempts(examId)
  if (attempts.value.length) {
    attemptId.value = attempts.value[0].attempt_id
    await loadDetail()
  }
})
</script>

<style scoped>
.point-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.point-name {
  font-size: 13px;
}
.point-total {
  font-size: 13px;
  color: var(--el-color-primary);
}
.point-total.overflow {
  color: var(--el-color-danger);
}
</style>
