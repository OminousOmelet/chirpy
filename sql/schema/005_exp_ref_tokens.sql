-- +goose Up
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE expires_at > NOW();

-- +goose Down
UPDATE refresh_tokens
SET revoked_at = NULL;