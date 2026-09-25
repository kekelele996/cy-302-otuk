<template>
  <div class="page-card">
    <div class="page-header">
      <h2>主观题批改</h2>
      <el-button @click="$router.push('/exams')">返回</el-button>
    </div>

    <div class="toolbar">
      <span>选择答题记录：</span>
      <el-select v-model="attemptId" placeholder="请选择" style="width: 360px" @change="loadDetail">
        <el-option v-for="a in attempts" :key="a.attempt_id" :label="`#${a.attempt_id} 客观分 ${a.objective_score}`" :value="a.attempt_id" />
      </el-select>
    </div>

    <template v-if="detail">
      <div v-if="!subjectiveQuestions.length" class="empty-wrap">
        <el-empty description="该试卷没有主观题" />
      </div>

      <div v-for="q in subjectiveQuestions" :key="q.exam_question_id" class="grade-card">
        <div class="grade-question">
          <el-tag size="small" type="info">{{ typeLabels[q.type] }}</el-tag>
          <span class="question-content">{{ q.content }}</span>
        </div>
        <el-row :gutter="16">
          <el-col :span="12">
            <div class="answer-block">
              <div class="answer-label">学生答案</div>
              <div class="answer-text">{{ formatAnswer(q.student_answer) }}</div>
            </div>
            <div class="answer-block">
              <div class="answer-label">参考答案</div>
              <div class="answer-text">{{ formatAnswer(q.correct_answer) }}</div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="answer-label">录分</div>
            <template v-if="hasPoints(q)">
              <div v-for="(point, idx) in q.scoring_points" :key="idx" class="point-row">
                <span class="point-name" :title="point.name">{{ point.name }}</span>
                <span class="point-max">满分 {{ point.score }}</span>
                <el-input-number
                  v-model="pointScores[q.exam_question_id][idx]"
                  :min="0"
                  :max="point.score"
                  :step="0.5"
                  size="small"
                />
              </div>
              <div class="point-total" :class="{ 'total-error': questionTotal(q) > q.max_score + 1e-6 }">
                本题合计：{{ questionTotal(q) }} / {{ q.max_score }}
              </div>
            </template>
            <template v-else>
              <div class="point-row legacy-row">
                <span class="point-name">整题给分（旧题无评分点）</span>
                <span class="point-max">满分 {{ q.max_score }}</span>
                <el-input-number v-model="scores[q.exam_question_id]" :min="0" :max="q.max_score" :step="0.5" size="small" />
              </div>
            </template>
          </el-col>
        </el-row>
      </div>

      <el-button
        v-if="subjectiveQuestions.length"
        type="primary"
        style="margin-top: 16px"
        :loading="saving"
        @click="onSubmit"
      >
        保存批改
      </el-button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { attemptApi, examApi } from '../api'
import type { AttemptDetail, AttemptQuestionDetail, AttemptSummary } from '../types'

const typeLabels: Record<string, string> = {
  fill_blank: '填空题',
  short_answer: '简答题'
}

const route = useRoute()
const examId = Number(route.params.examId)
const attempts = ref<AttemptSummary[]>([])
const attemptId = ref<number | null>(null)
const detail = ref<AttemptDetail | null>(null)
const scores = reactive<Record<number, number>>({})
const pointScores = reactive<Record<number, number[]>>({})
const saving = ref(false)

const subjectiveQuestions = computed(() => {
  if (!detail.value) return []
  return detail.value.questions.filter((q) => q.type === 'fill_blank' || q.type === 'short_answer')
})

function hasPoints(q: AttemptQuestionDetail) {
  return !!(q.scoring_points && q.scoring_points.length)
}

function questionTotal(q: AttemptQuestionDetail) {
  const list = pointScores[q.exam_question_id] || []
  return Math.round(list.reduce((sum, v) => sum + (Number(v) || 0), 0) * 100) / 100
}

function formatAnswer(v: unknown) {
  if (v === null || v === undefined) return '-'
  if (Array.isArray(v)) return v.join('；')
  return String(v)
}

async function loadDetail() {
  if (!attemptId.value) return
  detail.value = await attemptApi.detail(attemptId.value)
  for (const q of detail.value.questions) {
    if (q.type !== 'fill_blank' && q.type !== 'short_answer') continue
    if (hasPoints(q)) {
      pointScores[q.exam_question_id] = (q.scoring_points || []).map((p) =>
        q.graded && p.earned !== null && p.earned !== undefined ? p.earned : 0
      )
    } else {
      scores[q.exam_question_id] = q.graded ? q.score : 0
    }
  }
}

function validateBeforeSubmit() {
  for (const q of subjectiveQuestions.value) {
    if (hasPoints(q)) {
      const list = pointScores[q.exam_question_id] || []
      if (list.length !== q.scoring_points!.length) {
        ElMessage.error('评分点录分不完整')
        return false
      }
      for (let i = 0; i < list.length; i++) {
        const v = Number(list[i])
        if (!(v >= 0) || v > q.scoring_points![i].score + 1e-6) {
          ElMessage.error(`「${q.scoring_points![i].name}」得分不能超过 ${q.scoring_points![i].score} 分`)
          return false
        }
      }
      if (questionTotal(q) > q.max_score + 1e-6) {
        ElMessage.error(`本题合计不能超过 ${q.max_score} 分`)
        return false
      }
    } else if (Number(scores[q.exam_question_id]) > q.max_score + 1e-6) {
      ElMessage.error(`本题得分不能超过 ${q.max_score} 分`)
      return false
    }
  }
  return true
}

async function onSubmit() {
  if (!attemptId.value || !validateBeforeSubmit()) return
  saving.value = true
  try {
    const items = subjectiveQuestions.value.map((q) => {
      if (hasPoints(q)) {
        const list = pointScores[q.exam_question_id].map((v) => Number(v))
        return {
          exam_question_id: q.exam_question_id,
          score: list.reduce((sum, v) => sum + v, 0),
          point_scores: list
        }
      }
      return {
        exam_question_id: q.exam_question_id,
        score: Number(scores[q.exam_question_id]) || 0
      }
    })
    await attemptApi.grade(attemptId.value, items)
    ElMessage.success('批改成功')
    await loadDetail()
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
.empty-wrap {
  margin-top: 24px;
}
.grade-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 16px;
}
.grade-question {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  margin-bottom: 12px;
}
.question-content {
  line-height: 1.5;
}
.answer-block {
  margin-bottom: 10px;
}
.answer-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}
.answer-text {
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  padding: 6px 8px;
  min-height: 32px;
}
.point-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.point-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.point-max {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}
.legacy-row {
  padding-top: 2px;
}
.point-total {
  font-weight: 600;
  margin-top: 4px;
}
.total-error {
  color: var(--el-color-danger);
}
</style>
