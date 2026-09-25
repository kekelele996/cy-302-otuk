<template>
  <div class="page-card">
    <div class="page-header">
      <h2>成绩报告</h2>
      <el-button @click="$router.push('/attempts')">返回考试记录</el-button>
    </div>

    <template v-if="report">
      <el-descriptions :column="3" border>
        <el-descriptions-item label="考试">{{ report.exam_title }}</el-descriptions-item>
        <el-descriptions-item label="总分">{{ report.total_score }}</el-descriptions-item>
        <el-descriptions-item label="客观题得分">{{ report.objective_score }}</el-descriptions-item>
        <el-descriptions-item label="主观题得分">{{ report.subjective_score }}</el-descriptions-item>
        <el-descriptions-item label="客观题正确率">{{ report.accuracy }}%</el-descriptions-item>
        <el-descriptions-item label="排名">{{ report.rank }} / {{ report.participants }}</el-descriptions-item>
      </el-descriptions>

      <el-row :gutter="16" style="margin-top: 16px">
        <el-col :span="12">
          <div ref="chartRef" style="height: 320px"></div>
        </el-col>
        <el-col :span="12">
          <el-table :data="report.type_breakdown" border>
            <el-table-column prop="name" label="题型" />
            <el-table-column label="得分" width="120">
              <template #default="{ row }">{{ row.score }} / {{ row.max }}</template>
            </el-table-column>
            <el-table-column prop="count" label="题数" width="80" />
          </el-table>
        </el-col>
      </el-row>

      <template v-if="report.question_results && report.question_results.length">
        <h3 style="margin-top: 24px">主观题评分点明细</h3>
        <el-table :data="flattenedResults" border>
          <el-table-column label="题目" min-width="200">
            <template #default="{ row }">
              <div class="result-question">{{ row.content }}</div>
              <div class="result-meta">
                {{ typeLabels[row.type] }}
                <el-tag size="small" :type="row.graded ? 'success' : 'warning'">
                  {{ row.graded ? '已批改' : '待批' }}
                </el-tag>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="评分点" min-width="180">
            <template #default="{ row }">{{ row.point.name }}</template>
          </el-table-column>
          <el-table-column label="得分 / 满分" width="130" align="center">
            <template #default="{ row }">
              <span v-if="row.graded && row.point.earned !== null && row.point.earned !== undefined">
                <span :class="{ 'lost-score': row.point.earned < row.point.score }">{{ row.point.earned }}</span>
                / {{ row.point.score }}
              </span>
              <span v-else-if="row.graded">
                0 / {{ row.point.score }}
              </span>
              <el-tag v-else size="small" type="warning">待批</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="失分" width="100" align="center">
            <template #default="{ row }">
              <span v-if="row.graded && row.point.earned !== null && row.point.earned !== undefined">
                <span :class="{ 'lost-score': row.point.earned < row.point.score }">
                  -{{ Math.round((row.point.score - row.point.earned) * 100) / 100 }}
                </span>
              </span>
              <span v-else-if="row.graded">-{{ row.point.score }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { attemptApi } from '../api'
import type { ReportQuestionResult, ReportResponse, ScoringPointView } from '../types'

const typeLabels: Record<string, string> = {
  fill_blank: '填空题',
  short_answer: '简答题'
}

const route = useRoute()
const report = ref<ReportResponse | null>(null)
const chartRef = ref<HTMLElement | null>(null)

// Legacy subjective questions have no per-point rubric: synthesize a single
// point covering the whole question so the table still shows score / 待批.
const flattenedResults = computed(() => {
  if (!report.value) return []
  const rows: {
    exam_question_id: number
    type: string
    content: string
    graded: boolean
    point: ScoringPointView
  }[] = []
  for (const q of report.value.question_results as ReportQuestionResult[]) {
    if (q.points && q.points.length) {
      for (const point of q.points) {
        rows.push({ exam_question_id: q.exam_question_id, type: q.type, content: q.content, graded: q.graded, point })
      }
    } else {
      rows.push({
        exam_question_id: q.exam_question_id,
        type: q.type,
        content: q.content,
        graded: q.graded,
        point: { name: '整题得分', score: q.max_score, earned: q.graded ? q.score : null }
      })
    }
  }
  return rows
})

onMounted(async () => {
  const attemptId = Number(route.params.attemptId)
  report.value = await attemptApi.report(attemptId)
  await nextTick()
  renderChart()
})

function renderChart() {
  if (!chartRef.value || !report.value) return
  const chart = echarts.init(chartRef.value)
  const data = report.value.type_breakdown.map((item) => ({
    name: item.name,
    value: item.score
  }))
  chart.setOption({
    title: { text: '各题型得分分布', left: 'center' },
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [
      {
        type: 'pie',
        radius: ['35%', '65%'],
        data,
        label: { formatter: '{b}: {c}分' }
      }
    ]
  })
}
</script>

<style scoped>
.result-question {
  white-space: pre-wrap;
  word-break: break-word;
}
.result-meta {
  margin-top: 4px;
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.lost-score {
  color: var(--el-color-danger);
  font-weight: 600;
}
</style>
