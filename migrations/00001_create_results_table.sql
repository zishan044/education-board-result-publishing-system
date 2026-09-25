-- +goose Up
CREATE TABLE results (
    exam             TEXT         NOT NULL,
    exam_year        SMALLINT     NOT NULL,
    board            TEXT         NOT NULL,
    roll             INTEGER      NOT NULL,
    registration     BIGINT       NOT NULL,
    name             TEXT         NOT NULL,
    division         TEXT         NOT NULL,
    district         TEXT         NOT NULL,
    institution_code TEXT         NOT NULL,
    gpa              NUMERIC(3,2) NOT NULL,
    passed           BOOLEAN      NOT NULL,
    subjects         JSONB        NOT NULL
);

-- +goose Down
DROP TABLE results;