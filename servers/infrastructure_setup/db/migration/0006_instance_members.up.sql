BEGIN;

-- Who each resource was provisioned for, and whether they were granted access. An
-- instance only names its target (a team or one student), so without this a student
-- cannot be told which of their team's resources they can reach, and the participants
-- list cannot say which members a partial run left out. Rows are replaced on every run
-- of the instance, so they describe the members of its latest run.
CREATE TABLE resource_instance_member (
    resource_instance_id    uuid    NOT NULL REFERENCES resource_instance(id) ON DELETE CASCADE,
    course_participation_id uuid    NOT NULL,
    granted                 boolean NOT NULL,
    PRIMARY KEY (resource_instance_id, course_participation_id)
);

-- The student view and privacy requests look a person up across instances.
CREATE INDEX idx_resource_instance_member_participation
    ON resource_instance_member (course_participation_id);

COMMIT;
