-- Migration: Add course mail campaign tables

CREATE TYPE mail_campaign_status AS ENUM ('draft', 'sending', 'sent', 'partially_failed', 'failed');
CREATE TYPE mail_recipient_status AS ENUM ('pending', 'sent', 'failed');

-- Composite-FK targets so a campaign's target phase and a recipient's participation
-- can be pinned to the same course, not just validated at the service layer.
ALTER TABLE course_phase ADD CONSTRAINT unique_course_phase_id_course_id UNIQUE (id, course_id);
ALTER TABLE course_participation ADD CONSTRAINT unique_course_participation_id_course_id UNIQUE (id, course_id);

CREATE TABLE mail_campaign (
  id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id              uuid NOT NULL REFERENCES course(id) ON DELETE CASCADE,
  name                   text NOT NULL,
  subject                text NOT NULL DEFAULT '',
  body                   text NOT NULL DEFAULT '',
  target_course_phase_id uuid,
  target_pass_statuses   text[] NOT NULL DEFAULT '{}',
  reply_to_override      jsonb,
  cc_override            jsonb,
  bcc_override           jsonb,
  status                 mail_campaign_status NOT NULL DEFAULT 'draft',
  created_at             timestamptz NOT NULL DEFAULT now(),
  created_by_id          text NOT NULL,
  created_by_email       text NOT NULL DEFAULT '',
  created_by_name        text NOT NULL DEFAULT '',
  updated_at             timestamptz NOT NULL DEFAULT now(),
  updated_by_id          text NOT NULL DEFAULT '',
  updated_by_email       text NOT NULL DEFAULT '',
  updated_by_name        text NOT NULL DEFAULT '',
  sent_at                timestamptz,
  sent_by_id             text,
  sent_by_email          text,
  sent_by_name           text,
  UNIQUE (id, course_id),
  CONSTRAINT fk_mail_campaign_target_phase
    FOREIGN KEY (target_course_phase_id, course_id)
    REFERENCES course_phase(id, course_id)
    ON DELETE SET NULL (target_course_phase_id)
);

CREATE TABLE mail_campaign_recipient (
  id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id             uuid NOT NULL,
  course_id               uuid NOT NULL,
  course_participation_id uuid NOT NULL,
  email                   text NOT NULL,
  status                  mail_recipient_status NOT NULL DEFAULT 'pending',
  error_message           text NOT NULL DEFAULT '',
  sent_at                 timestamptz,
  UNIQUE (campaign_id, course_participation_id),
  CONSTRAINT fk_mail_campaign_recipient_campaign
    FOREIGN KEY (campaign_id, course_id)
    REFERENCES mail_campaign(id, course_id)
    ON DELETE CASCADE,
  CONSTRAINT fk_mail_campaign_recipient_participation
    FOREIGN KEY (course_participation_id, course_id)
    REFERENCES course_participation(id, course_id)
    ON DELETE CASCADE
);

-- The UNIQUE (campaign_id, course_participation_id) index already covers lookups and
-- the cascade by campaign_id; the participation cascade needs its own index.
CREATE INDEX idx_mail_campaign_course_id               ON mail_campaign(course_id);
CREATE INDEX idx_mail_campaign_recipient_participation ON mail_campaign_recipient(course_participation_id, course_id);
