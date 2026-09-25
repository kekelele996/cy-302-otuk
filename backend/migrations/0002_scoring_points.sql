-- Migration 0002: scoring points (grading rubric) support.
-- questions.scoring_points: rubric defined by the teacher while authoring a
--   fill_blank / short_answer question (JSON: [{"name","score"}]).
-- exam_questions.scoring_points: immutable rubric snapshot taken at paper
--   generation time, scaled to the paper per-question score. Grading always
--   uses this snapshot so editing a question later never changes old scores.
-- answers.point_scores: per-rubric awarded scores as a JSON array aligned
--   with the snapshot. Empty means legacy whole-question grading.

ALTER TABLE questions
    ADD COLUMN scoring_points TEXT NULL AFTER score;

ALTER TABLE exam_questions
    ADD COLUMN scoring_points TEXT NULL AFTER score;

ALTER TABLE answers
    ADD COLUMN point_scores TEXT NULL AFTER score;
