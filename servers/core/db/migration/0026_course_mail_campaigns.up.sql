-- Migration: Add course mail campaign tables

CREATE TYPE mail_campaign_status AS ENUM ('draft', 'sending', 'sent', 'partially_failed', 'failed');
CREATE TYPE mail_recipient_status AS ENUM ('pending', 'sent', 'failed');

CREATE TABLE mail_campaign (
  id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id              uuid NOT NULL REFERENCES course(id) ON DELETE CASCADE,
  name                   text NOT NULL,
  subject                text NOT NULL DEFAULT '',
  body                   text NOT NULL DEFAULT '',
  target_course_phase_id uuid REFERENCES course_phase(id) ON DELETE SET NULL,
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
  sent_by_name           text
);

CREATE TABLE mail_campaign_recipient (
  id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id             uuid NOT NULL REFERENCES mail_campaign(id) ON DELETE CASCADE,
  course_participation_id uuid NOT NULL,
  email                   text NOT NULL,
  status                  mail_recipient_status NOT NULL DEFAULT 'pending',
  error_message           text NOT NULL DEFAULT '',
  sent_at                 timestamptz,
  UNIQUE (campaign_id, course_participation_id)
);

CREATE INDEX idx_mail_campaign_course_id          ON mail_campaign(course_id);
CREATE INDEX idx_mail_campaign_recipient_campaign ON mail_campaign_recipient(campaign_id);
