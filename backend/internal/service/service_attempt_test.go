package service

import (
	"testing"

	"github.com/gbexam/online-exam/internal/dto"
	"github.com/gbexam/online-exam/internal/model"
)

func TestIsCorrectObjective(t *testing.T) {
	tests := []struct {
		name    string
		qtype   string
		correct any
		student any
		want    bool
	}{
		{"single correct", "single", "A", "A", true},
		{"single wrong", "single", "A", "B", false},
		{"true false case insensitive", "true_false", "T", "t", true},
		{"true false wrong", "true_false", "T", "F", false},
		{"multiple exact set", "multiple", []any{"A", "C"}, []any{"C", "A"}, true},
		{"multiple subset", "multiple", []any{"A", "C"}, []any{"A"}, false},
		{"multiple extra", "multiple", []any{"A"}, []any{"A", "B"}, false},
		{"unknown type never correct", "short_answer", "anything", "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCorrectObjective(tt.qtype, tt.correct, tt.student); got != tt.want {
				t.Fatalf("isCorrectObjective(%q, %v, %v) = %v, want %v", tt.qtype, tt.correct, tt.student, got, tt.want)
			}
		})
	}
}

func TestGradeByRubric(t *testing.T) {
	rubric := []dto.RubricPoint{
		{Name: "要点一", Score: 2},
		{Name: "要点二", Score: 3},
	}

	t.Run("sums awarded points", func(t *testing.T) {
		total, raw, err := gradeByRubric(rubric, []dto.GradePointRequest{
			{Name: "要点一", Score: 1.5},
			{Name: "要点二", Score: 3},
		}, 5)
		if err != nil {
			t.Fatalf("gradeByRubric() error = %v", err)
		}
		if total != 4.5 {
			t.Fatalf("gradeByRubric() total = %v, want 4.5", total)
		}
		awarded, err := unmarshalRubric(raw)
		if err != nil {
			t.Fatalf("unmarshal awarded: %v", err)
		}
		if len(awarded) != 2 || awarded[0].Score != 1.5 || awarded[1].Score != 3 {
			t.Fatalf("gradeByRubric() awarded = %+v", awarded)
		}
	})

	t.Run("rejects point score above its max", func(t *testing.T) {
		if _, _, err := gradeByRubric(rubric, []dto.GradePointRequest{
			{Name: "要点一", Score: 2.5},
			{Name: "要点二", Score: 3},
		}, 5); err == nil {
			t.Fatal("gradeByRubric() expected error for point score above max")
		}
	})

	t.Run("rejects missing point", func(t *testing.T) {
		if _, _, err := gradeByRubric(rubric, []dto.GradePointRequest{
			{Name: "要点一", Score: 2},
		}, 5); err == nil {
			t.Fatal("gradeByRubric() expected error for missing point")
		}
	})

	t.Run("rejects unknown point", func(t *testing.T) {
		if _, _, err := gradeByRubric(rubric, []dto.GradePointRequest{
			{Name: "要点一", Score: 2},
			{Name: "要点二", Score: 3},
			{Name: "要点三", Score: 1},
		}, 5); err == nil {
			t.Fatal("gradeByRubric() expected error for unknown point")
		}
	})

	t.Run("rejects negative point score", func(t *testing.T) {
		if _, _, err := gradeByRubric(rubric, []dto.GradePointRequest{
			{Name: "要点一", Score: -1},
			{Name: "要点二", Score: 3},
		}, 5); err == nil {
			t.Fatal("gradeByRubric() expected error for negative score")
		}
	})

	t.Run("rejects total above question max", func(t *testing.T) {
		uneven := []dto.RubricPoint{
			{Name: "要点一", Score: 2},
			{Name: "要点二", Score: 2},
		}
		if _, _, err := gradeByRubric(uneven, []dto.GradePointRequest{
			{Name: "要点一", Score: 2},
			{Name: "要点二", Score: 2},
		}, 3); err == nil {
			t.Fatal("gradeByRubric() expected error for total above question max")
		}
	})
}

func TestSnapshotRubric(t *testing.T) {
	rubricRaw, err := marshalRubric([]dto.RubricPoint{
		{Name: "要点一", Score: 1},
		{Name: "要点二", Score: 2},
	})
	if err != nil {
		t.Fatalf("marshalRubric() error = %v", err)
	}
	q := model.Question{Score: 3, Rubric: rubricRaw}

	t.Run("scales points to paper score", func(t *testing.T) {
		raw := snapshotRubric(q, 10)
		points, err := unmarshalRubric(raw)
		if err != nil {
			t.Fatalf("unmarshalRubric() error = %v", err)
		}
		if len(points) != 2 {
			t.Fatalf("snapshotRubric() returned %d points, want 2", len(points))
		}
		sum := points[0].Score + points[1].Score
		if sum != 10 {
			t.Fatalf("snapshotRubric() points sum = %v, want 10", sum)
		}
		if points[0].Name != "要点一" || points[1].Name != "要点二" {
			t.Fatalf("snapshotRubric() names = %+v", points)
		}
	})

	t.Run("keeps sum exact when scaling is uneven", func(t *testing.T) {
		raw := snapshotRubric(q, 1)
		points, _ := unmarshalRubric(raw)
		sum := 0.0
		for _, p := range points {
			sum += p.Score
		}
		if sum != 1 {
			t.Fatalf("snapshotRubric() points sum = %v, want 1", sum)
		}
	})

	t.Run("empty rubric yields empty snapshot", func(t *testing.T) {
		if raw := snapshotRubric(model.Question{Score: 5}, 5); raw != "" {
			t.Fatalf("snapshotRubric() = %q, want empty", raw)
		}
	})
}
