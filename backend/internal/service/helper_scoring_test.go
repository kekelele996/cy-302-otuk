package service

import (
	"math"
	"testing"

	"github.com/gbexam/online-exam/internal/constants"
	"github.com/gbexam/online-exam/internal/dto"
)

func TestNormalizeScoringPoints(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.QuestionRequest
		wantErr bool
	}{
		{
			name: "objective type ignores points",
			req: dto.QuestionRequest{
				Type:          constants.QuestionSingle,
				Score:         2,
				ScoringPoints: []dto.ScoringPoint{{Name: "要点", Score: 2}},
			},
			wantErr: false,
		},
		{
			name: "no points allowed for subjective legacy question",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 5,
			},
			wantErr: false,
		},
		{
			name: "valid points sum to score",
			req: dto.QuestionRequest{
				Type:  constants.QuestionFillBlank,
				Score: 4,
				ScoringPoints: []dto.ScoringPoint{
					{Name: "第一空", Score: 2},
					{Name: "第二空", Score: 2},
				},
			},
			wantErr: false,
		},
		{
			name: "blank name rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 4,
				ScoringPoints: []dto.ScoringPoint{
					{Name: "  ", Score: 2},
					{Name: "要点二", Score: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate name rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 4,
				ScoringPoints: []dto.ScoringPoint{
					{Name: "要点", Score: 2},
					{Name: "要点", Score: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "non-positive score rejected",
			req: dto.QuestionRequest{
				Type:          constants.QuestionShortAnswer,
				Score:         4,
				ScoringPoints: []dto.ScoringPoint{{Name: "要点", Score: 0}},
			},
			wantErr: true,
		},
		{
			name: "sum mismatch rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 6,
				ScoringPoints: []dto.ScoringPoint{
					{Name: "要点一", Score: 2},
					{Name: "要点二", Score: 3},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeScoringPoints(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeScoringPoints() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotScoringPoints(t *testing.T) {
	points := []dto.ScoringPoint{
		{Name: "要点一", Score: 3},
		{Name: "要点二", Score: 2},
	}

	t.Run("same score keeps values", func(t *testing.T) {
		got := snapshotScoringPoints(points, 5, 5)
		if len(got) != 2 || got[0].Score != 3 || got[1].Score != 2 {
			t.Fatalf("unexpected snapshot: %+v", got)
		}
	})

	t.Run("scaled scores sum to paper score", func(t *testing.T) {
		got := snapshotScoringPoints(points, 5, 10)
		sum := 0.0
		for _, p := range got {
			if p.Score <= 0 {
				t.Fatalf("scaled point must stay positive: %+v", got)
			}
			sum += p.Score
		}
		if math.Abs(sum-10) > 1e-9 {
			t.Fatalf("scaled snapshot sums to %.2f, want 10", sum)
		}
	})

	t.Run("scale down preserves names and order", func(t *testing.T) {
		got := snapshotScoringPoints(points, 5, 2)
		if got[0].Name != "要点一" || got[1].Name != "要点二" {
			t.Fatalf("names/order changed: %+v", got)
		}
		sum := 0.0
		for _, p := range got {
			sum += p.Score
		}
		if math.Abs(sum-2) > 1e-9 {
			t.Fatalf("scaled snapshot sums to %.2f, want 2", sum)
		}
	})

	t.Run("empty points stay empty", func(t *testing.T) {
		if got := snapshotScoringPoints(nil, 5, 10); got != nil {
			t.Fatalf("expected nil, got %+v", got)
		}
	})
}
