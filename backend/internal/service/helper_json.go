package service

import (
	"encoding/json"
	"fmt"

	"github.com/gbexam/online-exam/internal/dto"
)

func marshalOptions(options []dto.Option) (string, error) {
	if options == nil {
		options = []dto.Option{}
	}
	b, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("marshal options: %w", err)
	}
	return string(b), nil
}

func unmarshalOptions(raw string) ([]dto.Option, error) {
	if raw == "" {
		return []dto.Option{}, nil
	}
	var options []dto.Option
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil, fmt.Errorf("unmarshal options: %w", err)
	}
	return options, nil
}

func marshalAnswer(answer any) (string, error) {
	b, err := json.Marshal(answer)
	if err != nil {
		return "", fmt.Errorf("marshal answer: %w", err)
	}
	return string(b), nil
}

func unmarshalAnswer(raw string) (any, error) {
	if raw == "" {
		return nil, nil
	}
	var answer any
	if err := json.Unmarshal([]byte(raw), &answer); err != nil {
		return nil, fmt.Errorf("unmarshal answer: %w", err)
	}
	return answer, nil
}

// marshalRubric serializes rubric points; nil becomes an empty list.
func marshalRubric(points []dto.RubricPoint) (string, error) {
	if points == nil {
		points = []dto.RubricPoint{}
	}
	b, err := json.Marshal(points)
	if err != nil {
		return "", fmt.Errorf("marshal rubric: %w", err)
	}
	return string(b), nil
}

// unmarshalRubric parses stored rubric points; empty input yields an empty list.
func unmarshalRubric(raw string) ([]dto.RubricPoint, error) {
	if raw == "" {
		return []dto.RubricPoint{}, nil
	}
	var points []dto.RubricPoint
	if err := json.Unmarshal([]byte(raw), &points); err != nil {
		return nil, fmt.Errorf("unmarshal rubric: %w", err)
	}
	return points, nil
}
