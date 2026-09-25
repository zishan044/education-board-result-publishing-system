-- +goose Up
SET LOCAL maintenance_work_mem = '2GB';
ALTER TABLE results ADD PRIMARY KEY (exam, exam_year, board, roll, registration);

-- +goose Down
ALTER TABLE results DROP CONSTRAINT results_pkey;