package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/gbexam/online-exam/internal/constants"
	"github.com/gbexam/online-exam/internal/dto"
	"github.com/gbexam/online-exam/internal/model"
	"github.com/gbexam/online-exam/internal/repository"
)

// fakeAttemptStore is an in-memory implementation of the repos used by
// AttemptService.Grade / Report, sufficient for grading-flow tests.
type fakeAttemptStore struct {
	exams      map[uint]*model.Exam
	questions  map[uint]*model.Question
	examQs     map[uint][]model.ExamQuestion
	attempts   map[uint]*model.ExamAttempt
	answers    map[uint][]model.Answer
	nextAnswer uint
}

func newFakeAttemptStore() *fakeAttemptStore {
	return &fakeAttemptStore{
		exams:     map[uint]*model.Exam{},
		questions: map[uint]*model.Question{},
		examQs:    map[uint][]model.ExamQuestion{},
		attempts:  map[uint]*model.ExamAttempt{},
		answers:   map[uint][]model.Answer{},
	}
}

// --- ExamRepo ---

func (f *fakeAttemptStore) CreateExam(context.Context, *model.Exam) error { return nil }
func (f *fakeAttemptStore) FindExamByID(_ context.Context, id uint) (*model.Exam, error) {
	if e, ok := f.exams[id]; ok {
		copy := *e
		return &copy, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeAttemptStore) UpdateExam(context.Context, *model.Exam) error { return nil }
func (f *fakeAttemptStore) DeleteExam(context.Context, uint) error        { return nil }
func (f *fakeAttemptStore) ListExams(context.Context, repository.ExamFilter, int, int) ([]model.Exam, int64, error) {
	return nil, 0, nil
}
func (f *fakeAttemptStore) ReplaceExamQuestions(context.Context, uint, []model.ExamQuestion) error {
	return nil
}
func (f *fakeAttemptStore) ListExamQuestions(_ context.Context, examID uint) ([]model.ExamQuestion, error) {
	return append([]model.ExamQuestion(nil), f.examQs[examID]...), nil
}
func (f *fakeAttemptStore) CountExamQuestions(context.Context, uint) (int64, error) {
	return 0, nil
}

// --- QuestionRepo ---

func (f *fakeAttemptStore) CreateQuestion(context.Context, *model.Question) error { return nil }
func (f *fakeAttemptStore) CreateQuestionsBatch(context.Context, []model.Question) error {
	return nil
}
func (f *fakeAttemptStore) FindQuestionByID(_ context.Context, id uint) (*model.Question, error) {
	if q, ok := f.questions[id]; ok {
		copy := *q
		return &copy, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeAttemptStore) UpdateQuestion(context.Context, *model.Question) error { return nil }
func (f *fakeAttemptStore) DeleteQuestion(context.Context, uint) error            { return nil }
func (f *fakeAttemptStore) ListQuestions(context.Context, repository.QuestionFilter, int, int) ([]model.Question, int64, error) {
	return nil, 0, nil
}
func (f *fakeAttemptStore) ListQuestionsByTypeDifficulty(context.Context, string, string) ([]model.Question, error) {
	return nil, nil
}
func (f *fakeAttemptStore) FindQuestionsByIDs(_ context.Context, ids []uint) (map[uint]model.Question, error) {
	out := make(map[uint]model.Question, len(ids))
	for _, id := range ids {
		if q, ok := f.questions[id]; ok {
			out[id] = *q
		}
	}
	return out, nil
}

// --- AttemptRepo ---

func (f *fakeAttemptStore) CreateAttempt(context.Context, *model.ExamAttempt) error { return nil }
func (f *fakeAttemptStore) FindAttemptByID(_ context.Context, id uint) (*model.ExamAttempt, error) {
	if a, ok := f.attempts[id]; ok {
		copy := *a
		return &copy, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeAttemptStore) UpdateAttempt(_ context.Context, a *model.ExamAttempt) error {
	f.attempts[a.ID] = a
	return nil
}
func (f *fakeAttemptStore) FindInProgressAttempt(context.Context, uint, uint) (*model.ExamAttempt, error) {
	return nil, repository.ErrNotFound
}
func (f *fakeAttemptStore) ListAttemptsByStudent(context.Context, uint, uint, int, int) ([]model.ExamAttempt, int64, error) {
	return nil, 0, nil
}
func (f *fakeAttemptStore) ListAttemptsByExam(_ context.Context, examID uint) ([]model.ExamAttempt, error) {
	var out []model.ExamAttempt
	for _, a := range f.attempts {
		if a.ExamID == examID {
			out = append(out, *a)
		}
	}
	return out, nil
}

// --- AnswerRepo ---

func (f *fakeAttemptStore) SaveAnswer(_ context.Context, answer *model.Answer) error {
	list := f.answers[answer.AttemptID]
	for i, a := range list {
		if a.ExamQuestionID == answer.ExamQuestionID {
			if answer.ID == 0 {
				answer.ID = a.ID
			}
			list[i] = *answer
			f.answers[answer.AttemptID] = list
			return nil
		}
	}
	f.nextAnswer++
	answer.ID = f.nextAnswer
	f.answers[answer.AttemptID] = append(list, *answer)
	return nil
}
func (f *fakeAttemptStore) ListAnswersByAttempt(_ context.Context, attemptID uint) ([]model.Answer, error) {
	return append([]model.Answer(nil), f.answers[attemptID]...), nil
}

// --- WrongRepo (not exercised by grading tests) ---

func (f *fakeAttemptStore) UpsertWrongQuestion(context.Context, *model.WrongQuestion) error { return nil }
func (f *fakeAttemptStore) ListWrongQuestions(context.Context, uint, string, int, int) ([]model.WrongQuestion, int64, error) {
	return nil, 0, nil
}
func (f *fakeAttemptStore) DeleteWrongQuestion(context.Context, uint, uint) error      { return nil }
func (f *fakeAttemptStore) MarkWrongQuestionResolved(context.Context, uint, uint) error { return nil }

func newGradingService(t *testing.T) (*AttemptService, *fakeAttemptStore) {
	t.Helper()
	store := newFakeAttemptStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewAttemptService(store, store, store, store, store, logger)
	return svc, store
}

// seedGradingPaper builds one exam with two subjective questions:
// q1 has a 4-point rubric, q2 is a legacy 3-point question without rubric.
func seedGradingPaper(store *fakeAttemptStore) (uint, uint, uint) {
	store.exams[1] = &model.Exam{ID: 1, Title: "测试卷", TotalScore: 7, Status: constants.ExamPublished, CreatedBy: 9}
	// Current bank version of q1 has been edited to 10 points; the paper
	// snapshot must keep the original 4-point rubric.
	store.questions[101] = &model.Question{ID: 101, Type: constants.QuestionFillBlank, Score: 10, ScoringPoints: `[{"name":"新要点","score":10}]`}
	store.questions[102] = &model.Question{ID: 102, Type: constants.QuestionShortAnswer, Score: 3}
	store.examQs[1] = []model.ExamQuestion{
		{ID: 1001, ExamID: 1, QuestionID: 101, Score: 4, ScoringPoints: `[{"name":"第一空","score":2},{"name":"第二空","score":2}]`, SortOrder: 0},
		{ID: 1002, ExamID: 1, QuestionID: 102, Score: 3, SortOrder: 1},
	}
	store.attempts[1] = &model.ExamAttempt{ID: 1, ExamID: 1, StudentID: 7, Status: constants.AttemptSubmitted, ObjectiveScore: 0, TotalScore: 0}
	store.answers[1] = []model.Answer{
		{AttemptID: 1, ExamQuestionID: 1001, QuestionID: 101, AnswerText: `"北京"`, PointScores: `[0,0]`},
		{AttemptID: 1, ExamQuestionID: 1002, QuestionID: 102, AnswerText: `"答得不好"`},
	}
	return 1, 1001, 1002
}

func TestGradeWithScoringPoints(t *testing.T) {
	svc, store := newGradingService(t)
	attemptID, eq1, eq2 := seedGradingPaper(store)

	req := dto.GradeRequest{Items: []dto.GradeItemRequest{
		{ExamQuestionID: eq1, Score: 3, PointScores: []float64{2, 1}},
		{ExamQuestionID: eq2, Score: 2},
	}}
	if err := svc.Grade(context.Background(), 9, constants.RoleTeacher, attemptID, req); err != nil {
		t.Fatalf("Grade() unexpected error: %v", err)
	}
	if got := store.attempts[attemptID].TotalScore; got != 5 {
		t.Fatalf("total score = %.1f, want 5", got)
	}
	var a1 model.Answer
	for _, a := range store.answers[attemptID] {
		if a.ExamQuestionID == eq1 {
			a1 = a
		}
	}
	if a1.PointScores != "[2,1]" || a1.GradedBy != 9 {
		t.Fatalf("unexpected graded answer: %+v", a1)
	}
}

func TestGradeRejectsOverPointLimit(t *testing.T) {
	svc, store := newGradingService(t)
	attemptID, eq1, _ := seedGradingPaper(store)

	// 2.5 exceeds the single-point max of 2 even though the sum matches.
	req := dto.GradeRequest{Items: []dto.GradeItemRequest{
		{ExamQuestionID: eq1, Score: 3.5, PointScores: []float64{2.5, 1}},
	}}
	if err := svc.Grade(context.Background(), 9, constants.RoleTeacher, attemptID, req); err == nil {
		t.Fatal("expected over-point-limit error, got nil")
	}
	if store.attempts[attemptID].TotalScore != 0 {
		t.Fatalf("total must stay 0 after rejected grade, got %.1f", store.attempts[attemptID].TotalScore)
	}
}

func TestGradeRejectsOverQuestionLimit(t *testing.T) {
	svc, store := newGradingService(t)
	attemptID, _, eq2 := seedGradingPaper(store)

	req := dto.GradeRequest{Items: []dto.GradeItemRequest{
		{ExamQuestionID: eq2, Score: 4},
	}}
	if err := svc.Grade(context.Background(), 9, constants.RoleTeacher, attemptID, req); err == nil {
		t.Fatal("expected legacy over-question-limit error, got nil")
	}
}

func TestGradeRejectsPointCountMismatch(t *testing.T) {
	svc, store := newGradingService(t)
	attemptID, eq1, _ := seedGradingPaper(store)

	req := dto.GradeRequest{Items: []dto.GradeItemRequest{
		{ExamQuestionID: eq1, Score: 2, PointScores: []float64{2}},
	}}
	if err := svc.Grade(context.Background(), 9, constants.RoleTeacher, attemptID, req); err == nil {
		t.Fatal("expected point-count mismatch error, got nil")
	}
}

func TestReportShowsPendingAndGradedPoints(t *testing.T) {
	svc, store := newGradingService(t)
	attemptID, eq1, eq2 := seedGradingPaper(store)

	// Before grading: points are pending.
	before, err := svc.Report(context.Background(), constants.RoleStudent, 7, attemptID)
	if err != nil {
		t.Fatalf("Report() unexpected error: %v", err)
	}
	if len(before.QuestionResults) != 2 {
		t.Fatalf("question results = %d, want 2", len(before.QuestionResults))
	}
	pending := before.QuestionResults[0]
	if pending.Graded || len(pending.Points) != 2 || pending.Points[0].Earned != nil {
		t.Fatalf("unexpected pending result: %+v", pending)
	}

	// Grade q1 only; q2 stays pending (legacy).
	err = svc.Grade(context.Background(), 9, constants.RoleTeacher, attemptID, dto.GradeRequest{
		Items: []dto.GradeItemRequest{{ExamQuestionID: eq1, Score: 3, PointScores: []float64{2, 1}}},
	})
	if err != nil {
		t.Fatalf("Grade() unexpected error: %v", err)
	}
	after, err := svc.Report(context.Background(), constants.RoleStudent, 7, attemptID)
	if err != nil {
		t.Fatalf("Report() unexpected error: %v", err)
	}
	var q1, q2 dto.ReportQuestionResult
	for _, r := range after.QuestionResults {
		switch r.ExamQuestionID {
		case eq1:
			q1 = r
		case eq2:
			q2 = r
		}
	}
	if !q1.Graded || *q1.Points[0].Earned != 2 || *q1.Points[1].Earned != 1 {
		t.Fatalf("graded breakdown wrong: %+v", q1)
	}
	if q2.Graded {
		t.Fatalf("legacy ungraded question must remain pending: %+v", q2)
	}
}
