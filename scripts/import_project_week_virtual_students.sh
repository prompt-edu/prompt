#!/usr/bin/env bash
set -euo pipefail

CORE_URL="${CORE_URL:-http://localhost:8080}"
STUDENTS_FILE="${STUDENTS_FILE:-/home/zby/prompt_edu/templates/team_allocation/project_week_1000plus.virtual-students.json}"
APPLICATION_PHASE_ID="${APPLICATION_PHASE_ID:-}"
COURSE_ID="${COURSE_ID:-}"
TOKEN="${TOKEN:-}"
DRY_RUN="${DRY_RUN:-false}"
ENROLL_TO_COURSE="${ENROLL_TO_COURSE:-false}"
DEFAULT_LOCATION_AVAILABILITY_JSON="${DEFAULT_LOCATION_AVAILABILITY_JSON:-[\"Munich Onsite\",\"Remote\"]}"
DEFAULT_ENGLISH_PROFICIENCY="${DEFAULT_ENGLISH_PROFICIENCY:-C1/C2}"
DEFAULT_GERMAN_PROFICIENCY_DE="${DEFAULT_GERMAN_PROFICIENCY_DE:-Native}"
DEFAULT_GERMAN_PROFICIENCY_OTHER="${DEFAULT_GERMAN_PROFICIENCY_OTHER:-}"
DIVERSIFY_PREFERRED_FIELDS="${DIVERSIFY_PREFERRED_FIELDS:-true}"
DIVERSIFY_LANGUAGE_PROFICIENCY="${DIVERSIFY_LANGUAGE_PROFICIENCY:-true}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

require_cmd curl
require_cmd jq

if [[ -z "$APPLICATION_PHASE_ID" ]]; then
  echo "APPLICATION_PHASE_ID is required" >&2
  exit 1
fi

if [[ -z "$TOKEN" ]]; then
  echo "TOKEN is required" >&2
  exit 1
fi

if [[ "$ENROLL_TO_COURSE" == "true" && -z "$COURSE_ID" ]]; then
  echo "COURSE_ID is required when ENROLL_TO_COURSE=true" >&2
  exit 1
fi

if [[ ! -f "$STUDENTS_FILE" ]]; then
  echo "Students file not found: $STUDENTS_FILE" >&2
  exit 1
fi

echo "$DEFAULT_LOCATION_AVAILABILITY_JSON" | jq empty >/dev/null 2>&1 || {
  echo "DEFAULT_LOCATION_AVAILABILITY_JSON must be valid JSON" >&2
  exit 1
}

FORM_JSON="$(curl -fsS \
  -H "Authorization: Bearer $TOKEN" \
  "$CORE_URL/api/applications/$APPLICATION_PHASE_ID/form")"

get_question_id() {
  local section="$1"
  local access_key="$2"
  jq -r --arg section "$section" --arg key "$access_key" '
    def access_key_value:
      if .accessKey == null then ""
      elif (.accessKey | type) == "object" then (.accessKey.String // .accessKey.string // "")
      else .accessKey
      end;

    .[$section][]? | select(access_key_value == $key) | .id
  ' <<<"$FORM_JSON" | head -n1
}

get_question_options() {
  local access_key="$1"
  jq -c --arg key "$access_key" '
    def access_key_value:
      if .accessKey == null then ""
      elif (.accessKey | type) == "object" then (.accessKey.String // .accessKey.string // "")
      else .accessKey
      end;

    (.questionsMultiSelect[]? | select(access_key_value == $key) | .options) // []
  ' <<<"$FORM_JSON" | head -n1
}

normalize_gender() {
  local raw="${1:-}"
  local lowered
  lowered="$(tr '[:upper:]' '[:lower:]' <<<"$raw")"
  case "$lowered" in
    male) echo "male" ;;
    female) echo "female" ;;
    diverse) echo "diverse" ;;
    *) echo "prefer_not_to_say" ;;
  esac
}

normalize_study_degree() {
  local raw="${1:-}"
  local lowered
  lowered="$(tr '[:upper:]' '[:lower:]' <<<"$raw")"
  case "$lowered" in
    master|masters|m.sc.|msc) echo "master" ;;
    bachelor|bachelors|b.sc.|bsc) echo "bachelor" ;;
    *) echo "master" ;;
  esac
}

normalize_matriculation_number() {
  local raw="${1:-}"
  local digits
  digits="$(tr -cd '0-9' <<<"$raw")"
  if [[ -z "$digits" ]]; then
    printf '0%07d' 0
    return
  fi
  printf '%08d' "$((10#$digits))"
}

generate_university_login() {
  local first_name="$1"
  local last_name="$2"
  local index="$3"
  local first_letters last_letters
  first_letters="$(tr '[:upper:]' '[:lower:]' <<<"$first_name" | tr -cd '[:alpha:]' | cut -c1-2)"
  last_letters="$(tr '[:upper:]' '[:lower:]' <<<"$last_name" | tr -cd '[:alpha:]' | cut -c1-3)"

  while [[ ${#first_letters} -lt 2 ]]; do
    first_letters="${first_letters}x"
  done
  while [[ ${#last_letters} -lt 3 ]]; do
    last_letters="${last_letters}x"
  done

  printf '%s%02d%s' "$first_letters" "$((index % 100))" "$last_letters"
}

filter_multiselect_answer() {
  local raw_json="$1"
  local options_json="$2"
  jq -cn --argjson raw "$raw_json" --argjson options "$options_json" '
    ($raw | map(select(. as $item | $options | index($item)))) as $filtered
    | if ($filtered | length) > 0 then $filtered
      elif ($options | length) > 0 then [$options[0]]
      else []
      end
  '
}

pick_single_option() {
  local desired="$1"
  local options_json="$2"
  jq -rn --arg desired "$desired" --argjson options "$options_json" '
    if ($desired != "" and ($options | index($desired)) != null) then $desired
    elif ($options | length) > 0 then $options[0]
    else ""
    end
  '
}

pick_option_by_index() {
  local options_json="$1"
  local index="$2"
  jq -rn --argjson options "$options_json" --argjson idx "$index" '
    if ($options | length) == 0 then ""
    else $options[($idx % ($options | length))]
    end
  '
}

build_diversified_field_answer() {
  local options_json="$1"
  local index="$2"
  jq -cn --argjson options "$options_json" --argjson idx "$index" '
    if ($options | length) == 0 then []
    elif ($options | length) == 1 then [$options[0]]
    elif ($options | length) == 2 then [$options[$idx % 2], $options[($idx + 1) % 2]]
    else
      [
        $options[$idx % ($options | length)],
        $options[($idx + 2) % ($options | length)],
        $options[($idx + 4) % ($options | length)]
      ] | unique
    end
  '
}

resolve_student_id_by_email() {
  local email="$1"
  local response encoded_email

  encoded_email="$(jq -rn --arg v "$email" '$v|@uri')"

  response="$(curl -fsS \
    -H "Authorization: Bearer $TOKEN" \
    "$CORE_URL/api/students/search/$encoded_email")" || return 1

  jq -r --arg email "$email" '
    .[] | select((.email // "") == $email) | .id
  ' <<<"$response" | head -n1
}

enroll_student_to_course() {
  local student_id="$1"
  local response_file status_code response_body

  response_file="$(mktemp)"
  status_code="$(
    curl -sS -o "$response_file" -w "%{http_code}" \
      -X POST "$CORE_URL/api/courses/$COURSE_ID/participations/enroll" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      --data "{\"studentID\":\"$student_id\"}"
  )"
  response_body="$(<"$response_file")"
  rm -f "$response_file"

  if [[ "$status_code" == "200" ]]; then
    return 0
  fi

  if grep -qiE "duplicate key|already exists|unique|constraint" <<<"$response_body"; then
    return 2
  fi

  return 1
}

QID_MOTIVATION="$(get_question_id "questionsText" "motivation")"
QID_REGISTRATION_NOTES="$(get_question_id "questionsText" "project_week_registration_notes")"
QID_LOCATION_DETAILS="$(get_question_id "questionsText" "location_availability_details")"
QID_AVAILABILITY_CONSTRAINTS="$(get_question_id "questionsText" "availability_constraints")"
QID_PREFERRED_FIELDS="$(get_question_id "questionsMultiSelect" "preferred_field_of_business")"
QID_LOCATION_AVAILABILITY="$(get_question_id "questionsMultiSelect" "location_availability")"
QID_ENGLISH_PROFICIENCY="$(get_question_id "questionsMultiSelect" "language_proficiency_english")"
QID_GERMAN_PROFICIENCY="$(get_question_id "questionsMultiSelect" "language_proficiency_german")"

PREFERRED_FIELD_OPTIONS_JSON="$(get_question_options "preferred_field_of_business")"
LOCATION_OPTIONS_JSON="$(get_question_options "location_availability")"
ENGLISH_OPTIONS_JSON="$(get_question_options "language_proficiency_english")"
GERMAN_OPTIONS_JSON="$(get_question_options "language_proficiency_german")"

if [[ -z "$QID_MOTIVATION" || -z "$QID_PREFERRED_FIELDS" || -z "$QID_LOCATION_AVAILABILITY" || -z "$QID_ENGLISH_PROFICIENCY" ]]; then
  echo "Could not resolve the required application question IDs from the current form." >&2
  exit 1
fi

success_count=0
skip_count=0
fail_count=0
enroll_success_count=0
enroll_skip_count=0
enroll_fail_count=0
index=0

while IFS= read -r student_json; do
  index=$((index + 1))

  first_name="$(jq -r '.firstname // ""' <<<"$student_json")"
  last_name="$(jq -r '.lastname // ""' <<<"$student_json")"
  email="$(jq -r '.email // ""' <<<"$student_json")"
  nationality="$(jq -r '.countryofbirth // "DE"' <<<"$student_json")"
  semester_raw="$(jq -r '.semester // "1"' <<<"$student_json")"
  study_program="$(jq -r '.currentfieldofstudy // "Project Week Test Program"' <<<"$student_json")"
  study_degree_raw="$(jq -r '.currentdegree // "Masters"' <<<"$student_json")"
  gender_raw="$(jq -r '.gender // ""' <<<"$student_json")"
  matriculation_raw="$(jq -r '.matrikelnummer // ""' <<<"$student_json")"
  motivation="$(jq -r '.motivation // "Interested in participating in Project Week."' <<<"$student_json")"
  preferred_fields_raw_json="$(jq -c '.preferredFieldOfBusiness // []' <<<"$student_json")"
  english_override="$(jq -r '.englishProficiency // ""' <<<"$student_json")"
  german_override="$(jq -r '.germanProficiency // ""' <<<"$student_json")"

  semester_digits="$(tr -cd '0-9' <<<"$semester_raw")"
  if [[ -z "$semester_digits" ]]; then
    semester_digits="1"
  fi

  gender="$(normalize_gender "$gender_raw")"
  study_degree="$(normalize_study_degree "$study_degree_raw")"
  matriculation_number="$(normalize_matriculation_number "$matriculation_raw")"
  university_login="$(generate_university_login "$first_name" "$last_name" "$index")"

  if [[ "$DIVERSIFY_PREFERRED_FIELDS" == "true" ]]; then
    preferred_fields_answer_json="$(build_diversified_field_answer "$PREFERRED_FIELD_OPTIONS_JSON" "$index")"
  else
    preferred_fields_answer_json="$(filter_multiselect_answer "$preferred_fields_raw_json" "$PREFERRED_FIELD_OPTIONS_JSON")"
  fi
  location_answer_json="$(filter_multiselect_answer "$DEFAULT_LOCATION_AVAILABILITY_JSON" "$LOCATION_OPTIONS_JSON")"
  if [[ -n "$english_override" ]]; then
    english_level="$(pick_single_option "$english_override" "$ENGLISH_OPTIONS_JSON")"
  elif [[ "$DIVERSIFY_LANGUAGE_PROFICIENCY" == "true" ]]; then
    english_level="$(pick_option_by_index "$ENGLISH_OPTIONS_JSON" "$index")"
  else
    english_level="$(pick_single_option "$DEFAULT_ENGLISH_PROFICIENCY" "$ENGLISH_OPTIONS_JSON")"
  fi

  if [[ -n "$german_override" ]]; then
    german_desired="$german_override"
  elif [[ "$nationality" == "DE" ]]; then
    german_desired="$DEFAULT_GERMAN_PROFICIENCY_DE"
  else
    german_desired="$DEFAULT_GERMAN_PROFICIENCY_OTHER"
  fi

  if [[ "$DIVERSIFY_LANGUAGE_PROFICIENCY" == "true" && -z "$german_override" && -z "$german_desired" ]]; then
    german_level="$(pick_option_by_index "$GERMAN_OPTIONS_JSON" "$index")"
  else
    german_level="$(pick_single_option "$german_desired" "$GERMAN_OPTIONS_JSON")"
  fi

  payload="$(
    jq -n \
      --arg firstName "$first_name" \
      --arg lastName "$last_name" \
      --arg email "$email" \
      --arg matriculationNumber "$matriculation_number" \
      --arg universityLogin "$university_login" \
      --arg gender "$gender" \
      --arg nationality "$nationality" \
      --arg studyDegree "$study_degree" \
      --arg studyProgram "$study_program" \
      --argjson currentSemester "$semester_digits" \
      --arg motivationQid "$QID_MOTIVATION" \
      --arg motivation "$motivation" \
      --arg registrationNotesQid "$QID_REGISTRATION_NOTES" \
      --arg locationDetailsQid "$QID_LOCATION_DETAILS" \
      --arg availabilityConstraintsQid "$QID_AVAILABILITY_CONSTRAINTS" \
      --arg preferredFieldsQid "$QID_PREFERRED_FIELDS" \
      --arg locationAvailabilityQid "$QID_LOCATION_AVAILABILITY" \
      --arg englishQid "$QID_ENGLISH_PROFICIENCY" \
      --arg germanQid "$QID_GERMAN_PROFICIENCY" \
      --arg englishLevel "$english_level" \
      --arg germanLevel "$german_level" \
      --argjson preferredFieldsAnswer "$preferred_fields_answer_json" \
      --argjson locationAnswer "$location_answer_json" \
      '{
        student: {
          firstName: $firstName,
          lastName: $lastName,
          email: $email,
          matriculationNumber: $matriculationNumber,
          universityLogin: $universityLogin,
          hasUniversityAccount: true,
          gender: $gender,
          nationality: $nationality,
          studyDegree: $studyDegree,
          studyProgram: $studyProgram,
          currentSemester: $currentSemester
        },
        answersText: [
          if $motivationQid != "" then {applicationQuestionID: $motivationQid, answer: $motivation} else empty end
        ],
        answersMultiSelect: [
          if $preferredFieldsQid != "" then {applicationQuestionID: $preferredFieldsQid, answer: $preferredFieldsAnswer} else empty end,
          if $locationAvailabilityQid != "" then {applicationQuestionID: $locationAvailabilityQid, answer: $locationAnswer} else empty end,
          if $englishQid != "" and $englishLevel != "" then {applicationQuestionID: $englishQid, answer: [$englishLevel]} else empty end,
          if $germanQid != "" and $germanLevel != "" then {applicationQuestionID: $germanQid, answer: [$germanLevel]} else empty end
        ],
        answersFileUpload: []
      }'
  )"

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "DRY RUN: would import $first_name $last_name <$email>"
    echo "$payload" | jq .
    if [[ "$ENROLL_TO_COURSE" == "true" ]]; then
      echo "DRY RUN: would enroll into course $COURSE_ID"
    fi
    continue
  fi

  response_file="$(mktemp)"
  status_code="$(
    curl -sS -o "$response_file" -w "%{http_code}" \
      -X POST "$CORE_URL/api/applications/$APPLICATION_PHASE_ID" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      --data "$payload"
  )"
  response_body="$(cat "$response_file")"
  rm -f "$response_file"

  if [[ "$status_code" == "201" ]]; then
    success_count=$((success_count + 1))
    echo "Imported: $first_name $last_name <$email>"
  elif [[ "$status_code" == "405" ]] && grep -qi "already applied" <<<"$response_body"; then
    skip_count=$((skip_count + 1))
    echo "Skipped already applied: $first_name $last_name <$email>"
  else
    fail_count=$((fail_count + 1))
    echo "Failed to import $first_name $last_name <$email> (HTTP $status_code)" >&2
    echo "$response_body" >&2
    continue
  fi

  if [[ "$ENROLL_TO_COURSE" == "true" ]]; then
    student_id="$(resolve_student_id_by_email "$email")"
    if [[ -z "$student_id" ]]; then
      enroll_fail_count=$((enroll_fail_count + 1))
      echo "Failed to resolve student ID for $first_name $last_name <$email>; skip enrollment." >&2
      continue
    fi

    if enroll_student_to_course "$student_id"; then
      enroll_success_count=$((enroll_success_count + 1))
      echo "Enrolled into course: $first_name $last_name <$email>"
    else
      enroll_result=$?
      if [[ "$enroll_result" -eq 2 ]]; then
        enroll_skip_count=$((enroll_skip_count + 1))
        echo "Skipped already enrolled: $first_name $last_name <$email>"
      else
        enroll_fail_count=$((enroll_fail_count + 1))
        echo "Failed to enroll into course: $first_name $last_name <$email>" >&2
      fi
    fi
  fi
done < <(jq -c '.[]' "$STUDENTS_FILE")

echo
echo "Summary:"
echo "  imported: $success_count"
echo "  skipped:  $skip_count"
echo "  failed:   $fail_count"

if [[ "$ENROLL_TO_COURSE" == "true" ]]; then
  echo
  echo "Enrollment Summary:"
  echo "  enrolled: $enroll_success_count"
  echo "  skipped:  $enroll_skip_count"
  echo "  failed:   $enroll_fail_count"
fi

if [[ "$fail_count" -gt 0 || "$enroll_fail_count" -gt 0 ]]; then
  exit 1
fi
