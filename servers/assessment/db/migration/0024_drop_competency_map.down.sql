BEGIN;

CREATE TABLE competency_map
(
    from_competency_id uuid NOT NULL,
    to_competency_id   uuid NOT NULL,
    FOREIGN KEY (from_competency_id) REFERENCES competency (id) ON DELETE CASCADE,
    FOREIGN KEY (to_competency_id) REFERENCES competency (id) ON DELETE CASCADE,
    PRIMARY KEY (from_competency_id, to_competency_id)
);

COMMIT;
