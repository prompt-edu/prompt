#!/usr/bin/env bash
set -euo pipefail

# One-off migration of interview reviews out of core into the interview service.
#
# Until the interview phase became independent, score/interviewer/interviewAnswers lived in
# core's course_phase_participation.restricted_data. They now live in the interview service's
# interview_review table. Run this once, after both services have applied their migrations
# and before lecturers use the phase again.
#
# Usage: CORE_DB_URL=... INTERVIEW_DB_URL=... scripts/backfill-interview-reviews.sh
#
# Both URLs are libpq connection strings, e.g.
#   postgres://prompt-postgres:secret@localhost:5432/prompt
#
# The script is idempotent: rows that already exist in interview_review are left untouched,
# so re-running it never overwrites a review written after the cutover.

: "${CORE_DB_URL:?CORE_DB_URL must be set}"
: "${INTERVIEW_DB_URL:?INTERVIEW_DB_URL must be set}"

DUMP_FILE="$(mktemp -t interview_reviews.XXXXXX)"
trap 'rm -f "$DUMP_FILE"' EXIT

echo "Extracting interview reviews from core..."
psql "$CORE_DB_URL" -v ON_ERROR_STOP=1 --quiet <<SQL
\copy (SELECT p.course_phase_id, p.course_participation_id, CASE WHEN jsonb_typeof(p.restricted_data -> 'score') = 'number' THEN round((p.restricted_data ->> 'score')::numeric)::int END, NULLIF(p.restricted_data ->> 'interviewer', ''), CASE WHEN jsonb_typeof(p.restricted_data -> 'interviewAnswers') = 'array' THEN p.restricted_data -> 'interviewAnswers' ELSE '[]'::jsonb END FROM course_phase_participation p JOIN course_phase cp ON cp.id = p.course_phase_id JOIN course_phase_type cpt ON cpt.id = cp.course_phase_type_id WHERE cpt.name = 'Interview' AND (p.restricted_data ? 'score' OR p.restricted_data ? 'interviewer' OR p.restricted_data ? 'interviewAnswers')) TO '$DUMP_FILE'
SQL

echo "Loading $(wc -l < "$DUMP_FILE" | tr -d ' ') review(s) into the interview service..."
psql "$INTERVIEW_DB_URL" -v ON_ERROR_STOP=1 --quiet <<SQL
BEGIN;

CREATE TEMP TABLE interview_review_backfill (
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    score integer,
    interviewer varchar(255),
    interview_answers jsonb NOT NULL
) ON COMMIT DROP;

\copy interview_review_backfill FROM '$DUMP_FILE'

INSERT INTO interview_review (course_phase_id, course_participation_id, score, interviewer, interview_answers)
SELECT course_phase_id, course_participation_id, score, interviewer, interview_answers
FROM interview_review_backfill
ON CONFLICT (course_phase_id, course_participation_id) DO NOTHING;

COMMIT;
SQL

echo "Backfill complete."
