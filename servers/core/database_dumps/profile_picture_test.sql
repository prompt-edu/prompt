INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects) VALUES
('dddddddd-0000-0000-0000-000000000001', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10);

-- Ada has a picture (uploaded with her university login), Grace has none, Alan has no login.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender) VALUES
('aaaaaaaa-0000-0000-0000-000000000001', 'Ada', 'Lovelace', 'ada@tum.de', '03700001', 'ab12cde', true, 'female'),
('aaaaaaaa-0000-0000-0000-000000000002', 'Grace', 'Hopper', 'grace@tum.de', '03700002', 'gh34ijk', true, 'female'),
('aaaaaaaa-0000-0000-0000-000000000003', 'Alan', 'Turing', 'alan@example.com', NULL, NULL, false, 'male');

INSERT INTO course_participation (id, course_id, student_id) VALUES
('cccccccc-0000-0000-0000-000000000001', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000001'),
('cccccccc-0000-0000-0000-000000000002', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000002'),
('cccccccc-0000-0000-0000-000000000003', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000003');

-- Ada's account and an instructor account without a university login.
INSERT INTO files (id, filename, original_filename, content_type, size_bytes, storage_key, uploaded_by_user_id) VALUES
('ffffffff-0000-0000-0000-000000000001', 'ada.jpg', 'profile-picture.jpg', 'image/jpeg', 2048,
 'profile-picture/bbbbbbbb-0000-0000-0000-000000000001/ada.jpg', 'bbbbbbbb-0000-0000-0000-000000000001'),
('ffffffff-0000-0000-0000-000000000002', 'instructor.jpg', 'profile-picture.jpg', 'image/jpeg', 2048,
 'profile-picture/bbbbbbbb-0000-0000-0000-000000000002/instructor.jpg', 'bbbbbbbb-0000-0000-0000-000000000002');

INSERT INTO profile_picture (user_id, university_login, file_id) VALUES
('bbbbbbbb-0000-0000-0000-000000000001', 'ab12cde', 'ffffffff-0000-0000-0000-000000000001'),
('bbbbbbbb-0000-0000-0000-000000000002', NULL, 'ffffffff-0000-0000-0000-000000000002');
