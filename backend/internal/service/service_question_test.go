package service

import (
	"testing"

	"github.com/gbexam/online-exam/internal/constants"
	"github.com/gbexam/online-exam/internal/dto"
)

func TestValidateQuestion(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.QuestionRequest
		wantErr bool
	}{
		{
			name: "valid single choice",
			req: dto.QuestionRequest{
				Type: constants.QuestionSingle,
				Content: "1+1=?",
				Options: []dto.Option{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
				Answer:  "B",
				Difficulty: constants.DifficultyEasy,
				KnowledgePoint: "数学",
				Score: 2,
			},
			wantErr: false,
		},
		{
			name: "single choice answer not in options",
			req: dto.QuestionRequest{
				Type: constants.QuestionSingle,
				Content: "1+1=?",
				Options: []dto.Option{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}},
				Answer:  "C",
				Difficulty: constants.DifficultyEasy,
				KnowledgePoint: "数学",
				Score: 2,
			},
			wantErr: true,
		},
		{
			name: "valid true false",
			req: dto.QuestionRequest{
				Type: constants.QuestionTrueFalse,
				Content: "地球是圆的",
				Answer: "T",
				Difficulty: constants.DifficultyEasy,
				KnowledgePoint: "常识",
				Score: 1,
			},
			wantErr: false,
		},
		{
			name: "true false invalid answer",
			req: dto.QuestionRequest{
				Type: constants.QuestionTrueFalse,
				Content: "地球是圆的",
				Answer: "yes",
				Difficulty: constants.DifficultyEasy,
				KnowledgePoint: "常识",
				Score: 1,
			},
			wantErr: true,
		},
		{
			name: "valid multiple choice",
			req: dto.QuestionRequest{
				Type: constants.QuestionMultiple,
				Content: "选择偶数",
				Options: []dto.Option{{Key: "A", Text: "1"}, {Key: "B", Text: "2"}, {Key: "C", Text: "4"}},
				Answer:  []any{"B", "C"},
				Difficulty: constants.DifficultyMedium,
				KnowledgePoint: "数学",
				Score: 3,
			},
			wantErr: false,
		},
		{
			name: "fill blank valid",
			req: dto.QuestionRequest{
				Type: constants.QuestionFillBlank,
				Content: "中国首都是____",
				Answer: []any{"北京"},
				Difficulty: constants.DifficultyEasy,
				KnowledgePoint: "地理",
				Score: 2,
			},
			wantErr: false,
		},
		{
			name: "short answer valid",
			req: dto.QuestionRequest{
				Type: constants.QuestionShortAnswer,
				Content: "简述 HTTP 状态码",
				Answer: "200 表示成功",
				Difficulty: constants.DifficultyHard,
				KnowledgePoint: "网络",
				Score: 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateQuestion(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateQuestion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRubric(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.QuestionRequest
		wantLen int
		wantErr bool
	}{
		{
			name: "valid rubric sums to score",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 5,
				Rubric: []dto.RubricPoint{
					{Name: "概念正确", Score: 2},
					{Name: "举例恰当", Score: 3},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "valid fill blank rubric",
			req: dto.QuestionRequest{
				Type:  constants.QuestionFillBlank,
				Score: 4,
				Rubric: []dto.RubricPoint{
					{Name: "第一空", Score: 2},
					{Name: "第二空", Score: 2},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "sum mismatch rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 5,
				Rubric: []dto.RubricPoint{
					{Name: "要点一", Score: 2},
					{Name: "要点二", Score: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "empty name rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 2,
				Rubric: []dto.RubricPoint{
					{Name: "  ", Score: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate name rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 4,
				Rubric: []dto.RubricPoint{
					{Name: "要点", Score: 2},
					{Name: "要点", Score: 2},
				},
			},
			wantErr: true,
		},
		{
			name: "non positive point score rejected",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 2,
				Rubric: []dto.RubricPoint{
					{Name: "要点", Score: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "no rubric falls back to whole question grading",
			req: dto.QuestionRequest{
				Type:  constants.QuestionShortAnswer,
				Score: 5,
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "objective question ignores rubric",
			req: dto.QuestionRequest{
				Type:  constants.QuestionSingle,
				Score: 2,
				Rubric: []dto.RubricPoint{
					{Name: "要点", Score: 2},
				},
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points, err := validateRubric(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateRubric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(points) != tt.wantLen {
				t.Fatalf("validateRubric() returned %d points, want %d", len(points), tt.wantLen)
			}
		})
	}
}
