-- name: CreateLinkVisit :one
INSERT INTO link_visits (link_id, ip, user_agent, referer, status)
VALUES ($1, $2, $3, NULLIF($4, ''), $5)
RETURNING id, link_id, ip, user_agent, referer, status, created_at;

-- name: ListLinkVisitsPaginated :many
SELECT id, link_id, ip, user_agent, referer, status, created_at
FROM link_visits
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLinkVisits :one
SELECT COUNT(*) FROM link_visits;

-- name: DeleteLinkVisit :execrows
DELETE FROM link_visits
WHERE id = $1;
