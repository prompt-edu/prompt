-- name: CreateCategory :exec
INSERT INTO category (id, name, short_name, description, weight, assessment_schema_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: CheckCategoryNameExists :one
-- Check if a category name already exists within a given assessment schema
SELECT EXISTS(
    SELECT 1 FROM category
    WHERE assessment_schema_id = $1 AND name = $2
);

-- name: GetCategory :one
SELECT *
FROM category
WHERE id = $1;

-- name: ListCategories :many
SELECT *
FROM category
ORDER BY name ASC;

-- name: UpdateCategory :exec
UPDATE category
SET name                  = $2,
    short_name            = $3,
    description           = $4,
    weight                = $5,
    assessment_schema_id  = $6
WHERE id = $1;

-- name: DeleteCategory :exec
DELETE
FROM category
WHERE id = $1;

-- name: GetCategoriesWithCompetencies :many
SELECT c.id,
       c.name,
       c.short_name,
       c.description,
       c.weight,
       COALESCE(
                       json_agg(
                       json_build_object(
                               'id',
                               cmp.id,
                               'categoryID',
                               cmp.category_id,
                               'name',
                               cmp.name,
                               'shortName',
                               cmp.short_name,
                               'description',
                               cmp.description,
                               'descriptionVeryBad',
                               cmp.description_very_bad,
                               'descriptionBad',
                               cmp.description_bad,
                               'descriptionOk',
                               cmp.description_ok,
                               'descriptionGood',
                               cmp.description_good,
                               'descriptionVeryGood',
                               cmp.description_very_good,
                               'weight',
                               cmp.weight
                       )
                               ) FILTER (
                           WHERE cmp.id IS NOT NULL
                           ),
                       '[]'
       )::json AS competencies
FROM category c
         LEFT JOIN competency cmp ON c.id = cmp.category_id
WHERE c.assessment_schema_id = $1
GROUP BY c.id, c.name, c.short_name, c.description, c.weight
ORDER BY c.name ASC;