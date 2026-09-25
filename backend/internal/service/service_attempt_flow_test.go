package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/gbexam/online-exam/internal/constants"
	"github.com/gbexam/online-exam/internal/dto"
	"github.com/gbexam/online-exam/internal/model"
	"github.com/gbexam/online-exam/internal/repository"
)

// fakeStore is an in-memory implementation of all repo ports used by the
// exam/attempt flow, so the rubric snapshot behaviour can be tested end to end.
type fakeStore struct {
	questions          map[uint]model.Question
	exams              map[uint]model.Exam
	examQuestions      map[uint][]model.ExamQuestion
	attempts           map[uint]model.ExamAttempt
	answers            map[uint][]model.Answer
	nextExamID         uint
	nextAttemptID      uint
	nextExamQuestionID uint
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		questions:          map[uint]model.Question{},
		exams:              map[uint]model.Exam{},
		examQuestions:      map[uint][]model.ExamQuestion{},
		attempts:           map[uint]model.ExamAttempt{},
		answers:            map[uint][]model.Answer{},
		nextExamID:         1,
		nextAttemptID:      1,
		nextExamQuestionID: 1,
	}
}

func (f *fakeStore) CreateExam(_ context.Context, exam *model.Exam) error {
	exam.ID = f.nextExamID
	f.nextExamID++
	f.exams[exam.ID] = *exam
	return nil
}

func (f *fakeStore) FindExamByID(_ context.Context, id uint) (*model.Exam, error) {
	exam, ok := f.exams[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &exam, nil
}

func (f *fakeStore) UpdateExam(_ context.Context, exam *model.Exam) error {
	if _, ok := f.exams[exam.ID]; !ok {
		return repository.ErrNotFound
	}
	f.exams[exam.ID] = *exam
	return nil
}

func (f *fakeStore) DeleteExam(_ context.Context, id uint) error {
	delete(f.exams, id)
	return nil
}

func (f *fakeStore) ListExams(_ context.Context, _ repository.ExamFilter, _, _ int) ([]model.Exam, int64, error) {
	items := make([]model.Exam, 0, len(f.exams))
	for _, e := range f.exams {
		items = append(items, e)
	}
	return items, int64(len(items)), nil
}

func (f *fakeStore) ReplaceExamQuestions(_ context.Context, examID uint, items []model.ExamQuestion) error {
	stored := make([]model.ExamQuestion, 0, len(items))
	for _, it := range items {
		it.ID = f.nextExamQuestionID
		f.nextExamQuestionID++
		it.ExamID = examID
		stored = append(stored, it)
	}
	f.examQuestions[examID] = stored
	return nil
}

func (f *fakeStore) ListExamQuestions(_ context.Context, examID uint) ([]model.ExamQuestion, error) {
	return f.examQuestions[examID], nil
}

func (f *fakeStore) CountExamQuestions(_ context.Context, examID uint) (int64, error) {
	return int64(len(f.examQuestions[examID])), nil
}

func (f *fakeStore) CreateQuestion(_ context.Context, q *model.Question) error {
	f.questions[q.ID] = *q
	return nil
}

func (f *fakeStore) CreateQuestionsBatch(_ context.Context, questions []model.Question) error {
	for _, q := range questions {
		f.questions[q.ID] = q
	}
	return nil
}

func (f *fakeStore) FindQuestionByID(_ context.Context, id uint) (*model.Question, error) {
	q, ok := f.questions[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &q, nil
}

func (f *fakeStore) UpdateQuestion(_ context.Context, q *model.Question) error {
	f.questions[q.ID] = *q
	return nil
}

func (f *fakeStore) DeleteQuestion(_ context.Context, id uint) error {
	delete(f.questions, id)
	return nil
}

func (f *fakeStore) ListQuestions(_ context.Context, _ repository.QuestionFilter, _, _ int) ([]model.Question, int64, error) {
	items := make([]model.Question, 0, len(f.questions))
	for _, q := range f.questions {
		items = append(items, q)
	}
	return items, int64(len(items)), nil
}

func (f *fakeStore) ListQuestionsByTypeDifficulty(_ context.Context, qtype, difficulty string) ([]model.Question, error) {
	items := make([]model.Question, 0)
	for _, q := range f.questions {
		if q.Type != qtype {
			continue
		}
		if difficulty != "" && q.Difficulty != difficulty {
			continue
		}
		items = append(items, q)
	}
	return items, nil
}

func (f *fakeStore) FindQuestionsByIDs(_ context.Context, ids []uint) (map[uint]model.Question, error) {
	result := make(map[uint]model.Question, len(ids))
	for _, id := range ids {
		if q, ok := f.questions[id]; ok {
			result[id] = q
		}
	}
	return result, nil
}

func (f *fakeStore) CreateAttempt(_ context.Context, a *model.ExamAttempt) error {
	a.ID = f.nextAttemptID
	f.nextAttemptID++
	f.attempts[a.ID] = *a
	return nil
}

func (f *fakeStore) FindAttemptByID(_ context.Context, id uint) (*model.ExamAttempt, error) {
	a, ok := f.attempts[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return &a, nil
}

func (f *fakeStore) UpdateAttempt(_ context.Context, a *model.ExamAttempt) error {
	if _, ok := f.attempts[a.ID]; !ok {
		return repository.ErrNotFound
	}
	f.attempts[a.ID] = *a
	return nil
}

func (f *fakeStore) FindInProgressAttempt(_ context.Context, examID, studentID uint) (*model.ExamAttempt, error) {
	for _, a := range f.attempts {
		if a.ExamID == examID && a.StudentID == studentID && a.Status == constants.AttemptInProgress {
			copy := a
			return &copy, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *fakeStore) ListAttemptsByStudent(_ context.Context, studentID, _ uint, _, _ int) ([]model.ExamAttempt, int64, error) {
	items := make([]model.ExamAttempt, 0)
	for _, a := range f.attempts {
		if a.StudentID == studentID {
			items = append(items, a)
		}
	}
	return items, int64(len(items)), nil
}

func (f *fakeStore) ListAttemptsByExam(_ context.Context, examID uint) ([]model.ExamAttempt, error) {
	items := make([]model.ExamAttempt, 0)
	for _, a := range f.attempts {
		if a.ExamID == examID {
			items = append(items, a)
		}
	}
	return items, nil
}

func (f *fakeStore) SaveAnswer(_ context.Context, answer *model.Answer) error {
	items := f.answers[answer.AttemptID]
	for i := range items {
		if items[i].ExamQuestionID == answer.ExamQuestionID {
			answer.ID = items[i].ID
			items[i] = *answer
			f.answers[answer.AttemptID] = items
			return nil
		}
	}
	answer.ID = uint(len(items) + 1)
	f.answers[answer.AttemptID] = append(items, *answer)
	return nil
}

func (f *fakeStore) ListAnswersByAttempt(_ context.Context, attemptID uint) ([]model.Answer, error) {
	return f.answers[attemptID], nil
}

func (f *fakeStore) UpsertWrongQuestion(_ context.Context, _ *model.WrongQuestion) error {
	return nil
}

func (f *fakeStore) ListWrongQuestions(_ context.Context, _ uint, _ string, _, _ int) ([]model.WrongQuestion, int64, error) {
	return nil, 0, nil
}

func (f *fakeStore) DeleteWrongQuestion(_ context.Context, _, _ uint) error {
	return nil
}

func (f *fakeStore) MarkWrongQuestionResolved(_ context.Context, _, _ uint) error {
	return nil
}

// TestRubricSnapshotFlow walks the whole lifecycle: create a rubric question,
// generate a paper (snapshot), edit the bank question, then take and grade the
// exam. Grading and reporting must follow the snapshot, not the edited question.
func TestRubricSnapshotFlow(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	logger := slog.Default()

	rubricRaw, err := marshalRubric([]dto.RubricPoint{
		{Name: "要点一", Score: 2},
		{Name: "要点二", Score: 3},
	})
	if err != nil {
		t.Fatalf("marshalRubric() error = %v", err)
	}
	store.questions[1] = model.Question{
		ID:             1,
		Type:           constants.QuestionShortAnswer,
		Content:        "简述缓存穿透",
		Answer:         `"参考：布隆过滤器、空值缓存"`,
		Difficulty:     constants.DifficultyEasy,
		KnowledgePoint: "缓存",
		Score:          5,
		Rubric:         rubricRaw,
	}

	examSvc := NewExamService(store, store, logger)
	attemptSvc := NewAttemptService(store, store, store, store, store, logger)

	// 组卷：试卷中该题 10 分，评分点应按比例快照
	exam, err := examSvc.Create(ctx, 1, dto.ExamCreateRequest{
		Title:           "期末考",
		DurationMinutes: 60,
		QuestionConfig:  []dto.PaperQuestionConfig{{Type: constants.QuestionShortAnswer, Count: 1, Score: 10}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	eqs, err := store.ListExamQuestions(ctx, exam.ID)
	if err != nil || len(eqs) != 1 {
		t.Fatalf("ListExamQuestions() = %v, %v", eqs, err)
	}
	snapshot, _ := unmarshalRubric(eqs[0].Rubric)
	if len(snapshot) != 2 || snapshot[0].Score != 4 || snapshot[1].Score != 6 {
		t.Fatalf("snapshot rubric = %+v, want 要点一=4 要点二=6", snapshot)
	}

	// 发布后再改题库评分点，旧试卷不应受影响
	if err := examSvc.Publish(ctx, constants.RoleAdmin, 1, exam.ID); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	edited := store.questions[1]
	edited.Rubric, _ = marshalRubric([]dto.RubricPoint{{Name: "新要点", Score: 5}})
	store.questions[1] = edited

	start, err := attemptSvc.Start(ctx, 100, exam.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	eqID := start.Questions[0].ExamQuestionID
	if err := attemptSvc.SaveAnswer(ctx, 100, start.AttemptID, dto.AnswerSubmitRequest{ExamQuestionID: eqID, Answer: "学生作答"}); err != nil {
		t.Fatalf("SaveAnswer() error = %v", err)
	}
	if err := attemptSvc.Submit(ctx, 100, start.AttemptID); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	// 未批改时报告显示待批（Graded=false）
	report, err := attemptSvc.Report(ctx, constants.RoleAdmin, 1, start.AttemptID)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if len(report.SubjectiveItems) != 1 {
		t.Fatalf("SubjectiveItems = %+v, want 1 item", report.SubjectiveItems)
	}
	item := report.SubjectiveItems[0]
	if item.Graded {
		t.Fatal("expected Graded=false before grading")
	}
	if len(item.Points) != 2 || item.Points[0].Max != 4 || item.Points[1].Max != 6 {
		t.Fatalf("report points = %+v, want snapshot maxes 4 and 6", item.Points)
	}

	// 超过快照要点满分必须拒绝
	err = attemptSvc.Grade(ctx, 1, constants.RoleAdmin, start.AttemptID, dto.GradeRequest{
		Items: []dto.GradeItemRequest{{
			ExamQuestionID: eqID,
			PointScores: []dto.GradePointRequest{
				{Name: "要点一", Score: 5},
				{Name: "要点二", Score: 6},
			},
		}},
	})
	if err == nil {
		t.Fatal("Grade() expected error when point score exceeds snapshot max")
	}

	// 按快照要点正常批改：4 + 3 = 7
	err = attemptSvc.Grade(ctx, 1, constants.RoleAdmin, start.AttemptID, dto.GradeRequest{
		Items: []dto.GradeItemRequest{{
			ExamQuestionID: eqID,
			PointScores: []dto.GradePointRequest{
				{Name: "要点一", Score: 4},
				{Name: "要点二", Score: 3},
			},
		}},
	})
	if err != nil {
		t.Fatalf("Grade() error = %v", err)
	}
	attempt, err := store.FindAttemptByID(ctx, start.AttemptID)
	if err != nil {
		t.Fatalf("FindAttemptByID() error = %v", err)
	}
	if attempt.TotalScore != 7 {
		t.Fatalf("TotalScore = %v, want 7", attempt.TotalScore)
	}

	// 报告按快照展示得分与失分，而不是改后的题库评分点
	report, err = attemptSvc.Report(ctx, constants.RoleAdmin, 1, start.AttemptID)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	item = report.SubjectiveItems[0]
	if !item.Graded || item.Score != 7 || item.MaxScore != 10 {
		t.Fatalf("graded item = %+v, want score 7 / 10", item)
	}
	if item.Points[0].Score != 4 || item.Points[0].Lost != 0 || item.Points[1].Score != 3 || item.Points[1].Lost != 3 {
		t.Fatalf("report points = %+v, want 4/4 and 3/6", item.Points)
	}
}

// TestGradeWithoutRubricKeepsWholeQuestionGrading covers legacy questions
// whose paper snapshot has no rubric: they are still graded as a whole.
func TestGradeWithoutRubricKeepsWholeQuestionGrading(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	logger := slog.Default()

	store.questions[1] = model.Question{
		ID:             1,
		Type:           constants.QuestionFillBlank,
		Content:        "首都是____",
		Answer:         `["北京"]`,
		Difficulty:     constants.DifficultyEasy,
		KnowledgePoint: "地理",
		Score:          4,
	}

	examSvc := NewExamService(store, store, logger)
	attemptSvc := NewAttemptService(store, store, store, store, store, logger)

	exam, err := examSvc.Create(ctx, 1, dto.ExamCreateRequest{
		Title:           "测验",
		DurationMinutes: 30,
		QuestionConfig:  []dto.PaperQuestionConfig{{Type: constants.QuestionFillBlank, Count: 1, Score: 4}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	eqs, _ := store.ListExamQuestions(ctx, exam.ID)
	if eqs[0].Rubric != "" {
		t.Fatalf("Rubric snapshot = %q, want empty for question without rubric", eqs[0].Rubric)
	}
	if err := examSvc.Publish(ctx, constants.RoleAdmin, 1, exam.ID); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	start, err := attemptSvc.Start(ctx, 100, exam.ID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	eqID := start.Questions[0].ExamQuestionID
	if err := attemptSvc.SaveAnswer(ctx, 100, start.AttemptID, dto.AnswerSubmitRequest{ExamQuestionID: eqID, Answer: []string{"北京"}}); err != nil {
		t.Fatalf("SaveAnswer() error = %v", err)
	}
	if err := attemptSvc.Submit(ctx, 100, start.AttemptID); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	// 整题给分超过题目满分必须拒绝
	err = attemptSvc.Grade(ctx, 1, constants.RoleAdmin, start.AttemptID, dto.GradeRequest{
		Items: []dto.GradeItemRequest{{ExamQuestionID: eqID, Score: 4.5}},
	})
	if err == nil {
		t.Fatal("Grade() expected error when whole-question score exceeds max")
	}

	if err := attemptSvc.Grade(ctx, 1, constants.RoleAdmin, start.AttemptID, dto.GradeRequest{
		Items: []dto.GradeItemRequest{{ExamQuestionID: eqID, Score: 3}},
	}); err != nil {
		t.Fatalf("Grade() error = %v", err)
	}
	attempt, _ := store.FindAttemptByID(ctx, start.AttemptID)
	if attempt.TotalScore != 3 {
		t.Fatalf("TotalScore = %v, want 3", attempt.TotalScore)
	}

	report, err := attemptSvc.Report(ctx, constants.RoleAdmin, 1, start.AttemptID)
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	item := report.SubjectiveItems[0]
	if !item.Graded || item.Score != 3 || len(item.Points) != 0 {
		t.Fatalf("report item = %+v, want graded whole-question score 3 without points", item)
	}
}
