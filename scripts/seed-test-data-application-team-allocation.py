#!/usr/bin/env python3
"""
Seed diverse test data for Team Allocation + TEASE.

Creates two courses:
  - "APTest-Small"  (SS26) with ~50 participants and 5 teams
  - "APTest-Large"  (SS26) with ~200 participants and 10 teams

Each course has an Application phase (initial) and a Team Allocation phase.
Team Allocation DB is seeded with survey timeframes, skills, teams, and random
student skill + team-preference responses so TEASE has data to render.

Idempotent: deletes any prior seed by fixed course UUIDs before re-inserting.

Run from the prompt/ directory:
    python scripts/seed-test-data-application-team-allocation.py
"""

import io
import random
import subprocess
import uuid
from datetime import datetime, timedelta

# --- Config ------------------------------------------------------------------

CORE_CONTAINER = "prompt-db"
TA_CONTAINER = "prompt-db-team-allocation"
DB_USER = "prompt-postgres"
DB_NAME = "prompt"

APP_PHASE_TYPE_ID = "935b16ee-4488-4a83-b417-6b202a0a2c7d"
TA_PHASE_TYPE_ID = "b6f6e709-d9b1-42a2-9afd-862028e6be4c"

# participation_data_dependency_graph DTO IDs
APP_OUT_APPLICATION_ANSWERS_DTO_ID = "fa759ace-ed7c-4197-b1b0-db41bb4a2234"
APP_OUT_SCORE_LEVEL_DTO_ID = "043bc029-83d7-4752-ba2a-99e53220e4aa"
TA_IN_APPLICATION_ANSWERS_DTO_ID = "81ca3c22-230e-4f3b-bdfd-6b8ca6eac8b7"
TA_IN_SCORE_LEVEL_DTO_ID = "f9ed9369-7373-45bd-96ca-aaea413fb359"

SEMESTER_TAG = "ss2026"

COURSE_SMALL = {
    "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "name": "APTest-Small",
    "students": 50,
    "teams": 5,
    "app_phase_id": "aaaaaaaa-aaaa-aaaa-aaaa-a00000000001",
    "ta_phase_id": "aaaaaaaa-aaaa-aaaa-aaaa-a00000000002",
}
COURSE_LARGE = {
    "id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
    "name": "APTest-Large",
    "students": 200,
    "teams": 10,
    "app_phase_id": "bbbbbbbb-bbbb-bbbb-bbbb-b00000000001",
    "ta_phase_id": "bbbbbbbb-bbbb-bbbb-bbbb-b00000000002",
}
COURSES = [COURSE_SMALL, COURSE_LARGE]

random.seed(42)

# --- Data pools for diverse students ----------------------------------------

FIRST_NAMES_MALE = [
    "Lukas", "Maximilian", "Jonas", "Finn", "Paul", "Leon", "Noah", "Elias",
    "Ben", "Liam", "Mateo", "Luca", "Hugo", "Mohammed", "Ahmed", "Aarav",
    "Raj", "Wei", "Hiroshi", "Diego", "Carlos", "Santiago", "Oliver", "Henry",
    "Jakob", "Felix", "Tim", "Tobias", "Marco", "Giulio", "Matteo", "Pietro",
    "Pierre", "Antoine", "Mikhail", "Alexei", "Dmitri", "Arjun", "Kenji",
    "Yusuf", "Omar", "Ali", "Kwame", "Jamal", "Ethan", "Mason", "Logan",
]
FIRST_NAMES_FEMALE = [
    "Sophia", "Emma", "Mia", "Hannah", "Anna", "Lea", "Lena", "Leonie", "Lina",
    "Amelia", "Olivia", "Ava", "Isabella", "Chloe", "Zoe", "Maya", "Aisha",
    "Fatima", "Layla", "Ananya", "Priya", "Mei", "Yuki", "Sakura", "Sofia",
    "Valentina", "Camila", "Elena", "Giulia", "Chiara", "Marta", "Beatrice",
    "Claire", "Margot", "Juliette", "Tatiana", "Nadia", "Svetlana", "Aarti",
    "Hina", "Zara", "Amina", "Nora", "Farida", "Adaeze", "Chidinma", "Grace",
]
FIRST_NAMES_NEUTRAL = ["Alex", "Sam", "Jordan", "Taylor", "Morgan", "Robin"]

LAST_NAMES = [
    "Müller", "Schmidt", "Schneider", "Fischer", "Weber", "Meyer", "Wagner",
    "Becker", "Schulz", "Hoffmann", "Schäfer", "Koch", "Bauer", "Richter",
    "Klein", "Wolf", "Schröder", "Neumann", "Schwarz", "Zimmermann",
    "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Martinez",
    "Rodriguez", "Hernandez", "Lopez", "Gonzalez", "Perez", "Sanchez",
    "Rossi", "Russo", "Ferrari", "Ricci", "Marino", "Greco", "Bruno",
    "Dubois", "Leroy", "Moreau", "Laurent", "Bernard", "Fontaine",
    "Ivanov", "Petrov", "Sokolov", "Volkov", "Kuznetsov", "Popov",
    "Patel", "Shah", "Kumar", "Singh", "Sharma", "Gupta", "Reddy",
    "Chen", "Wang", "Li", "Zhang", "Liu", "Yang", "Huang", "Zhao",
    "Tanaka", "Yamamoto", "Suzuki", "Kobayashi", "Sato", "Watanabe",
    "Ali", "Khan", "Hussein", "Rahman", "Karimov", "Abdi", "Okafor",
]

# (code, probability weight)
NATIONALITIES = [
    ("DE", 50), ("AT", 6), ("CH", 4), ("US", 6), ("GB", 4), ("FR", 5),
    ("IT", 5), ("ES", 3), ("NL", 2), ("PL", 3), ("IN", 4), ("CN", 4),
    ("JP", 2), ("BR", 2), ("MX", 2), ("TR", 2), ("EG", 1), ("NG", 1),
    ("KR", 1), ("SG", 1),
]

STUDY_PROGRAMS = [
    "Informatik", "Wirtschaftsinformatik", "Mathematik", "Maschinenwesen",
    "Elektrotechnik und Informationstechnik", "Data Engineering and Analytics",
    "Robotics, Cognition, Intelligence", "Informatik: Games Engineering",
    "Bioinformatik", "Medizinische Informatik", "Physik",
    "Management & Technology", "Finance & Information Management",
]

GENDERS_WEIGHTED = [("male", 50), ("female", 43), ("diverse", 5), ("prefer_not_to_say", 2)]
DEGREES_WEIGHTED = [("bachelor", 60), ("master", 40)]

SKILLS = [
    "React.js", "TypeScript", "Python", "C++", "Java", "Swift (iOS)",
    "Kotlin (Android)", "SQL & Databases", "Machine Learning",
    "DevOps & CI/CD", "UI/UX Design", "Agile & Scrum",
]

# Application-phase language questions. Each answer is exported via access_key
# so later phases (Team Allocation) can read the value.
LANGUAGE_OPTIONS = ["A1/A2", "B1/B2", "C1/C2", "Native"]
LANGUAGE_QUESTIONS = [
    {
        "title": "English Proficiency",
        "access_key": "language_proficiency_english",
        # Most students aren't native; weight toward B1/B2 and C1/C2.
        "weights": [10, 45, 35, 10],
    },
    {
        "title": "German Proficiency",
        "access_key": "language_proficiency_german",
        # Many internationals — broader spread, fewer natives than English.
        "weights": [25, 35, 25, 15],
    },
]

PROJECT_NAMES = [
    "BMW Autonomous Fleet Tracker",
    "Siemens IoT Monitoring Platform",
    "Allianz Claims Automation",
    "Deutsche Bahn Routing Optimizer",
    "Munich Re Risk Analytics Dashboard",
    "Airbus Maintenance Companion",
    "SAP Workflow Orchestrator",
    "Infineon Chip Defect Classifier",
    "Lufthansa Crew Scheduling Assistant",
    "TUM Campus AR Wayfinder",
    "MAN Truck Telematics Hub",
    "Continental Tire Wear Predictor",
]


# --- Helpers -----------------------------------------------------------------


def weighted_choice(pairs):
    """Pick a value from `pairs` ((value, weight) tuples) proportional to weight."""
    total = sum(w for _, w in pairs)
    r = random.uniform(0, total)
    upto = 0
    for val, w in pairs:
        upto += w
        if r <= upto:
            return val
    return pairs[-1][0]


def psql(container, sql):
    """Execute `sql` inside the named Docker `container` via psql and return stdout.

    Forces UTF-8 client encoding so non-ASCII names survive the shell round-trip.
    Raises RuntimeError on a non-zero exit (ON_ERROR_STOP=1).
    """
    # Force UTF-8 on both ends so names like "Müller" survive the shell.
    full_sql = "SET client_encoding = 'UTF8';\n" + sql
    result = subprocess.run(
        ["docker", "exec", "-i", container, "psql", "-U", DB_USER,
         "-d", DB_NAME, "-v", "ON_ERROR_STOP=1"],
        input=full_sql.encode("utf-8"), capture_output=True,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"psql failed on {container}:\n{result.stderr.decode('utf-8', errors='replace')}"
        )
    return result.stdout.decode("utf-8", errors="replace")


def sql_literal(v):
    """Render a Python value as a PostgreSQL literal suitable for inline SQL."""
    if v is None:
        return "NULL"
    if isinstance(v, bool):
        return "TRUE" if v else "FALSE"
    if isinstance(v, (int, float)):
        return str(v)
    if isinstance(v, uuid.UUID):
        return f"'{v}'"
    # treat as string
    return "'" + str(v).replace("'", "''") + "'"


def insert_many(container, table, columns, rows, chunk=500):
    """Bulk-insert `rows` into `table`/`columns` in the given container, in chunks."""
    if not rows:
        return
    cols = ", ".join(columns)
    for i in range(0, len(rows), chunk):
        part = rows[i:i + chunk]
        values = ",\n".join(
            "(" + ", ".join(sql_literal(v) for v in row) + ")"
            for row in part
        )
        psql(container, f"INSERT INTO {table} ({cols}) VALUES\n{values};\n")


# --- Students generator ------------------------------------------------------


def pick_first_name(gender):
    """Return a random first name from the pool matching `gender` (mixed for diverse/other)."""
    if gender == "male":
        return random.choice(FIRST_NAMES_MALE)
    if gender == "female":
        return random.choice(FIRST_NAMES_FEMALE)
    return random.choice(FIRST_NAMES_MALE + FIRST_NAMES_FEMALE + FIRST_NAMES_NEUTRAL)


def generate_student(idx):
    """Build a synthetic student record with unique identifiers derived from `idx`."""
    gender = weighted_choice(GENDERS_WEIGHTED)
    first = pick_first_name(gender)
    last = random.choice(LAST_NAMES)
    degree = weighted_choice(DEGREES_WEIGHTED)
    semester = random.randint(1, 6) if degree == "bachelor" else random.randint(1, 4)
    program = random.choice(STUDY_PROGRAMS)
    nationality = weighted_choice(NATIONALITIES)

    # Unique fields
    matric = f"{10000000 + idx:08d}"
    login = f"tst{idx:05d}"
    # Strip accents crudely for email
    email_first = "".join(c for c in first.lower() if c.isalpha())
    email_last = "".join(c for c in last.lower() if c.isalpha())
    email = f"{email_first}.{email_last}.{idx}@test.tum.de"

    return {
        "id": uuid.uuid4(),
        "first_name": first,
        "last_name": last,
        "email": email,
        "matriculation_number": matric,
        "university_login": login,
        "has_university_account": True,
        "gender": gender,
        "nationality": nationality,
        "study_program": program,
        "study_degree": degree,
        "current_semester": semester,
    }


# --- Seeding -----------------------------------------------------------------


def cleanup():
    """Remove any data previously inserted by this seeder across both databases."""
    course_ids = ", ".join(f"'{c['id']}'" for c in COURSES)
    ta_phase_ids = ", ".join(f"'{c['ta_phase_id']}'" for c in COURSES)

    # Team Allocation DB first (depends on course_participation IDs,
    # but phase_id is self-sufficient here)
    psql(TA_CONTAINER, f"""
        DELETE FROM allocations WHERE course_phase_id IN ({ta_phase_ids});
        DELETE FROM student_team_preference_response
            WHERE team_id IN (SELECT id FROM team WHERE course_phase_id IN ({ta_phase_ids}));
        DELETE FROM student_skill_response
            WHERE skill_id IN (SELECT id FROM skill WHERE course_phase_id IN ({ta_phase_ids}));
        DELETE FROM tutor WHERE course_phase_id IN ({ta_phase_ids});
        DELETE FROM team WHERE course_phase_id IN ({ta_phase_ids});
        DELETE FROM skill WHERE course_phase_id IN ({ta_phase_ids});
        DELETE FROM survey_timeframe WHERE course_phase_id IN ({ta_phase_ids});
    """)

    # Core DB — cascading deletes handle participations and phases
    psql(CORE_CONTAINER, f"""
        DELETE FROM student
            WHERE id IN (
                SELECT DISTINCT student_id FROM course_participation
                WHERE course_id IN ({course_ids})
            );
        DELETE FROM course WHERE id IN ({course_ids});
    """)


def seed_core(course, students):
    """Seed the core DB for `course`: course row, phases, participations, and application data."""
    # Course + phases. Team Allocation sits immediately after Application
    # via course_phase_graph.
    psql(CORE_CONTAINER, f"""
        INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, template)
        VALUES (
            '{course['id']}', '{course['name']}',
            DATE '2026-04-01', DATE '2026-09-30',
            '{SEMESTER_TAG}', 'practical course', 10, FALSE
        );

        INSERT INTO course_phase (id, course_id, name, is_initial_phase, course_phase_type_id)
        VALUES
            ('{course['app_phase_id']}', '{course['id']}', 'Application', TRUE, '{APP_PHASE_TYPE_ID}'),
            ('{course['ta_phase_id']}', '{course['id']}', 'Team Allocation', FALSE, '{TA_PHASE_TYPE_ID}');

        INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
        VALUES ('{course['app_phase_id']}', '{course['ta_phase_id']}');

        -- Wire Application -> Team Allocation data dependencies so that
        -- applicationAnswers (languages) and scoreLevel flow into PrevData.
        INSERT INTO participation_data_dependency_graph
            (from_course_phase_id, to_course_phase_id,
             from_course_phase_dto_id, to_course_phase_dto_id)
        VALUES
            ('{course['app_phase_id']}', '{course['ta_phase_id']}',
             '{APP_OUT_APPLICATION_ANSWERS_DTO_ID}', '{TA_IN_APPLICATION_ANSWERS_DTO_ID}'),
            ('{course['app_phase_id']}', '{course['ta_phase_id']}',
             '{APP_OUT_SCORE_LEVEL_DTO_ID}', '{TA_IN_SCORE_LEVEL_DTO_ID}');
    """)

    # Students
    student_cols = [
        "id", "first_name", "last_name", "email", "matriculation_number",
        "university_login", "has_university_account", "gender", "nationality",
        "study_program", "study_degree", "current_semester",
    ]
    student_rows = [[s[c] for c in student_cols] for s in students]
    insert_many(CORE_CONTAINER, "student", student_cols, student_rows)

    # Course participations (course × student)
    participations = []
    for s in students:
        cp_id = uuid.uuid4()
        s["course_participation_id"] = cp_id
        participations.append([cp_id, course["id"], s["id"]])
    insert_many(CORE_CONTAINER, "course_participation",
                ["id", "course_id", "student_id"], participations)

    # Course phase participations — application (passed) + team allocation (not_assessed)
    cpp_rows = []
    for s in students:
        cpp_rows.append([s["course_participation_id"], course["app_phase_id"], "passed"])
        cpp_rows.append([s["course_participation_id"], course["ta_phase_id"], "not_assessed"])
    insert_many(CORE_CONTAINER, "course_phase_participation",
                ["course_participation_id", "course_phase_id", "pass_status"], cpp_rows)

    # Language proficiency application questions (multi-select, single choice).
    # accessible_for_other_phases=true + access_key lets Team Allocation read
    # these answers by key.
    question_rows = []
    question_ids = []
    for order_num, q in enumerate(LANGUAGE_QUESTIONS, start=1):
        qid = uuid.uuid4()
        question_ids.append((qid, q))
        question_rows.append([
            qid,
            course["app_phase_id"],
            q["title"],
            f"Please select your current {q['title'].lower()} level (CEFR).",
            None,                   # placeholder
            "Please select a level.",
            True,                   # is_required
            1, 1,                   # min_select, max_select
            "{" + ",".join(f'"{opt}"' for opt in LANGUAGE_OPTIONS) + "}",
            order_num,
            True,                   # accessible_for_other_phases
            q["access_key"],
        ])
    insert_many(
        CORE_CONTAINER, "application_question_multi_select",
        ["id", "course_phase_id", "title", "description", "placeholder",
         "error_message", "is_required", "min_select", "max_select", "options",
         "order_num", "accessible_for_other_phases", "access_key"],
        question_rows,
    )

    # Student answers for each language question
    answer_rows = []
    for s in students:
        for qid, q in question_ids:
            choice = random.choices(LANGUAGE_OPTIONS, weights=q["weights"])[0]
            answer_rows.append([
                uuid.uuid4(),
                qid,
                "{" + f'"{choice}"' + "}",
                s["course_participation_id"],
            ])
    insert_many(
        CORE_CONTAINER, "application_answer_multi_select",
        ["id", "application_question_id", "answer", "course_participation_id"],
        answer_rows,
    )

    # Application assessment scores. 1 = veryGood, 5 = veryBad — weight toward the middle.
    # The Application phase emits this as `scoreLevel` to Team Allocation.
    score_weights = [10, 25, 35, 20, 10]  # for scores 1..5
    assessment_rows = []
    for s in students:
        score = random.choices([1, 2, 3, 4, 5], weights=score_weights)[0]
        assessment_rows.append([
            uuid.uuid4(),
            score,
            course["app_phase_id"],
            s["course_participation_id"],
        ])
    insert_many(
        CORE_CONTAINER, "application_assessment",
        ["id", "score", "course_phase_id", "course_participation_id"],
        assessment_rows,
    )


def seed_team_allocation(course, students):
    """Seed the team-allocation DB: survey window, skills, teams, and per-student responses."""
    phase_id = course["ta_phase_id"]

    # Survey timeframe — currently open
    start = datetime.utcnow() - timedelta(days=7)
    deadline = datetime.utcnow() + timedelta(days=7)
    psql(TA_CONTAINER, f"""
        INSERT INTO survey_timeframe (course_phase_id, survey_start, survey_deadline)
        VALUES ('{phase_id}',
                TIMESTAMP '{start.isoformat(timespec='seconds')}',
                TIMESTAMP '{deadline.isoformat(timespec='seconds')}');
    """)

    # Skills
    skill_rows = []
    skill_ids = []
    for name in SKILLS:
        sid = uuid.uuid4()
        skill_ids.append(sid)
        skill_rows.append([sid, phase_id, name])
    insert_many(TA_CONTAINER, "skill", ["id", "course_phase_id", "name"], skill_rows)

    # Teams (projects)
    team_rows = []
    team_ids = []
    chosen_projects = random.sample(PROJECT_NAMES, k=course["teams"])
    for name in chosen_projects:
        tid = uuid.uuid4()
        team_ids.append(tid)
        team_rows.append([tid, name, phase_id])
    insert_many(TA_CONTAINER, "team", ["id", "name", "course_phase_id"], team_rows)

    # Skill responses — give each student a "persona" bias so responses cluster
    levels = ["novice", "intermediate", "advanced", "expert"]
    skill_response_rows = []
    for s in students:
        # Pick 1-2 strong skills per student
        strong_idx = set(random.sample(range(len(skill_ids)), k=random.randint(1, 2)))
        for i, sid in enumerate(skill_ids):
            if i in strong_idx:
                level = random.choices(levels, weights=[0, 10, 40, 50])[0]
            else:
                level = random.choices(levels, weights=[30, 45, 20, 5])[0]
            skill_response_rows.append([s["course_participation_id"], sid, level])
    insert_many(TA_CONTAINER, "student_skill_response",
                ["course_participation_id", "skill_id", "skill_level"],
                skill_response_rows)

    # Team preferences — each student ranks all teams (1 = top)
    pref_rows = []
    for s in students:
        ranked = random.sample(team_ids, k=len(team_ids))
        for rank, tid in enumerate(ranked, start=1):
            pref_rows.append([s["course_participation_id"], tid, rank])
    insert_many(TA_CONTAINER, "student_team_preference_response",
                ["course_participation_id", "team_id", "preference"], pref_rows)

    # Allocations are intentionally left empty — allocation is what we're testing.


def main():
    """Entry point: clean up any prior seed, then seed both configured courses."""
    print("Cleaning up any prior seed…")
    cleanup()

    next_idx = 1
    for course in COURSES:
        print(f"Seeding {course['name']} ({course['students']} students, {course['teams']} teams)…")
        students = [generate_student(next_idx + i) for i in range(course["students"])]
        next_idx += course["students"]
        seed_core(course, students)
        seed_team_allocation(course, students)

    print("Done.")
    print("")
    print("Summary:")
    for c in COURSES:
        print(f"  - {c['name']}  course_id={c['id']}")
        print(f"      application phase    = {c['app_phase_id']}")
        print(f"      team allocation phase= {c['ta_phase_id']}")


if __name__ == "__main__":
    main()
