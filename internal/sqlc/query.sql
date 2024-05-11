-- name: GetCompany :many
SELECT * FROM companies
WHERE name = ?
ORDER BY site;

-- name: ListCompanies :many
SELECT * FROM companies
ORDER BY site;

-- name: AddCompany :exec
INSERT INTO companies (
    name, site
) VALUES (
    ?, ?
);
