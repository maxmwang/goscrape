-- name: GetCompanies :many
SELECT * FROM companies
ORDER BY site;

-- name: AddCompany :exec
INSERT INTO companies (
    name, site
) VALUES (
    ?, ?
);
