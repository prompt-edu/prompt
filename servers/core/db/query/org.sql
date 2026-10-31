-- name: CreateOrg :one
INSERT INTO org (id, parent_org_id, name, slug, school, university, website, contact_email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetOrgByID :one
SELECT * FROM org WHERE id = $1;

-- name: GetAllOrgs :many
SELECT * FROM org ORDER BY name ASC, slug ASC;

-- name: UpdateOrg :one
UPDATE org
SET name = $2, school = $3, university = $4, website = $5, contact_email = $6
WHERE id = $1
RETURNING *;

-- name: UpdateOrgParent :one
UPDATE org SET parent_org_id = $2 WHERE id = $1 RETURNING *;

-- name: GetOrgAncestorIDs :many
WITH RECURSIVE ancestors (id) AS (
  SELECT o.parent_org_id
  FROM org o
  WHERE o.id = $1 AND o.parent_org_id IS NOT NULL
  UNION
  SELECT o.parent_org_id
  FROM org o
  JOIN ancestors a ON o.id = a.id
  WHERE o.parent_org_id IS NOT NULL
)
SELECT id::uuid AS id FROM ancestors;

-- name: LockOrgHierarchy :exec
SELECT pg_advisory_xact_lock(hashtext('org_hierarchy'));

-- name: DeleteOrg :one
DELETE FROM org WHERE id = $1 RETURNING slug;
