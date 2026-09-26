-- One platform profile picture per Keycloak user. Students are matched through their university
-- login at read time, so a student row created after the upload still resolves to the picture.
CREATE TABLE IF NOT EXISTS profile_picture
(
    user_id          UUID PRIMARY KEY,
    university_login VARCHAR(20),
    file_id          UUID      NOT NULL UNIQUE,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_profile_picture_file FOREIGN KEY (file_id) REFERENCES files (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_profile_picture_university_login ON profile_picture (university_login);
