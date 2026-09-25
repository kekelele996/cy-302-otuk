-- Migration 0002: scoring points (rubric) for subjective questions.
-- questions.rubric stores the rubric defined on the bank question;
-- exam_questions.rubric is the snapshot frozen when the paper is generated;
-- answers.point_scores stores the awarded score per rubric point.
-- The Go server also runs GORM AutoMigrate at startup, which adds these
-- columns automatically; apply this script manually only if AutoMigrate is
-- not used.

ALTER TABLE questions ADD COLUMN rubric TEXT NULL AFTER score;
ALTER TABLE exam_questions ADD COLUMN rubric TEXT NULL AFTER score;
ALTER TABLE answers ADD COLUMN point_scores TEXT NULL AFTER score;
