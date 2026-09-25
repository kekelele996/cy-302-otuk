package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gbexam/online-exam/internal/constants"
	"github.com/gbexam/online-exam/internal/dto"
)

// scoreEpsilon tolerates floating point drift when summing point scores.
const scoreEpsilon = 1e-6

// supportsScoringPoints reports whether a question type is graded by teachers
// through scoring points (fill_blank / short_answer).
func supportsScoringPoints(qtype string) bool {
	return qtype == constants.QuestionFillBlank || qtype == constants.QuestionShortAnswer
}

// normalizeScoringPoints validates the rubric of a question request.
//
// Points are only meaningful for fill_blank / short_answer questions and are
// ignored for objective types. For subjective questions points are optional:
// a question without points keeps legacy whole-question grading. When points
// are present, every name must be non-empty and unique, every score must be
// positive, and the scores must sum to the question total.
func normalizeScoringPoints(req dto.QuestionRequest) ([]dto.ScoringPoint, error) {
	if !supportsScoringPoints(req.Type) || len(req.ScoringPoints) == 0 {
		return nil, nil
	}
	points := make([]dto.ScoringPoint, 0, len(req.ScoringPoints))
	names := map[string]bool{}
	var sum float64
	for _, p := range req.ScoringPoints {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return nil, fmt.Errorf("%w: 评分点名称不能为空", ErrValidation)
		}
		if names[name] {
			return nil, fmt.Errorf("%w: 评分点名称重复: %s", ErrValidation, name)
		}
		names[name] = true
		if p.Score <= 0 {
			return nil, fmt.Errorf("%w: 评分点「%s」分值必须大于 0", ErrValidation, name)
		}
		points = append(points, dto.ScoringPoint{Name: name, Score: roundScore(p.Score)})
		sum += p.Score
	}
	if math.Abs(roundScore(sum)-req.Score) > scoreEpsilon {
		return nil, fmt.Errorf("%w: 评分点分值之和 %.2f 必须等于题目总分 %.2f", ErrValidation, roundScore(sum), req.Score)
	}
	return points, nil
}

// snapshotScoringPoints scales the question-bank rubric to the per-question
// score assigned by paper generation. The scaled points always sum exactly
// to paperScore (largest-remainder allocation at cent granularity), so later
// rubric edits cannot change how an old paper is graded.
func snapshotScoringPoints(points []dto.ScoringPoint, questionScore, paperScore float64) []dto.ScoringPoint {
	if len(points) == 0 {
		return nil
	}
	if math.Abs(questionScore-paperScore) <= scoreEpsilon {
		return append([]dto.ScoringPoint(nil), points...)
	}
	type allocation struct {
		index int
		value int // hundredths
		floor int
		frac  float64
	}
	totalCents := int(math.Round(paperScore * 100))
	allocations := make([]allocation, len(points))
	floorSum := 0
	for i, p := range points {
		scaled := p.Score / questionScore * paperScore * 100
		floor := int(scaled)
		allocations[i] = allocation{index: i, floor: floor, frac: scaled - float64(floor)}
		floorSum += floor
	}
	remainder := totalCents - floorSum
	sort.Slice(allocations, func(i, j int) bool {
		if allocations[i].frac != allocations[j].frac {
			return allocations[i].frac > allocations[j].frac
		}
		return allocations[i].index < allocations[j].index
	})
	for i := 0; i < remainder; i++ {
		allocations[i].value = allocations[i].floor + 1
	}
	for i := remainder; i < len(allocations); i++ {
		allocations[i].value = allocations[i].floor
	}
	snapshot := make([]dto.ScoringPoint, len(points))
	for _, a := range allocations {
		snapshot[a.index] = dto.ScoringPoint{Name: points[a.index].Name, Score: float64(a.value) / 100}
	}
	return snapshot
}

func marshalScoringPoints(points []dto.ScoringPoint) (string, error) {
	if len(points) == 0 {
		return "", nil
	}
	b, err := json.Marshal(points)
	if err != nil {
		return "", fmt.Errorf("marshal scoring points: %w", err)
	}
	return string(b), nil
}

// unmarshalScoringPoints parses a stored rubric snapshot. Empty input means
// the question (or paper item) has no rubric and uses whole-question grading.
func unmarshalScoringPoints(raw string) ([]dto.ScoringPoint, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var points []dto.ScoringPoint
	if err := json.Unmarshal([]byte(raw), &points); err != nil {
		return nil, fmt.Errorf("unmarshal scoring points: %w", err)
	}
	return points, nil
}

func marshalPointScores(scores []float64) (string, error) {
	if len(scores) == 0 {
		return "", nil
	}
	b, err := json.Marshal(scores)
	if err != nil {
		return "", fmt.Errorf("marshal point scores: %w", err)
	}
	return string(b), nil
}

func unmarshalPointScores(raw string) []float64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var scores []float64
	if err := json.Unmarshal([]byte(raw), &scores); err != nil {
		return nil
	}
	return scores
}

func roundScore(v float64) float64 {
	return math.Round(v*100) / 100
}
