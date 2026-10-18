package postgres

import (
	"context"
	"qmanga/internal/identity/application"
)

func (r *Repository) SaveAvatar(ctx context.Context, a application.Avatar) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO avatars(owner_kind,owner_id,object_key,bytes,updated_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(owner_kind,owner_id) DO UPDATE SET object_key=excluded.object_key,bytes=excluded.bytes,updated_at=excluded.updated_at`, a.Kind, a.OwnerID, a.Key, a.Bytes, a.UpdatedAt)
	return translate(err)
}
func (r *Repository) Avatar(ctx context.Context, kind, owner string) (application.Avatar, error) {
	a := application.Avatar{Kind: kind, OwnerID: owner}
	err := r.pool.QueryRow(ctx, `SELECT a.object_key,a.bytes,a.updated_at FROM avatars a WHERE a.owner_kind=$1 AND a.owner_id=$2 AND (($1='user' AND EXISTS(SELECT 1 FROM users u WHERE u.id=a.owner_id AND NOT u.is_blocked)) OR ($1='team' AND EXISTS(SELECT 1 FROM teams t WHERE t.id=a.owner_id AND t.is_approved)))`, kind, owner).Scan(&a.Key, &a.Bytes, &a.UpdatedAt)
	return a, translate(err)
}
