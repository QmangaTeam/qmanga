package postgres

import (
	"context"
	"qmanga/internal/identity/application"
	"time"
)

func (r *Repository) ReserveAvatarObject(ctx context.Context, key string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO avatar_garbage(object_key,ready_at) VALUES($1,now()+interval '1 hour')`, key)
	return err
}
func (r *Repository) AbandonAvatarObject(ctx context.Context, key string) error {
	_, err := r.pool.Exec(ctx, `UPDATE avatar_garbage SET ready_at=now() WHERE object_key=$1 AND NOT EXISTS(SELECT 1 FROM avatars WHERE object_key=$1)`, key)
	return err
}
func (r *Repository) AvatarGarbage(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT g.object_key FROM avatar_garbage g WHERE ready_at<now() AND NOT EXISTS(SELECT 1 FROM avatars a WHERE a.object_key=g.object_key) ORDER BY ready_at LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}
func (r *Repository) ForgetAvatarGarbage(ctx context.Context, key string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM avatar_garbage WHERE object_key=$1 AND NOT EXISTS(SELECT 1 FROM avatars WHERE object_key=$1)`, key)
	return err
}
func (r *Repository) SaveAvatar(ctx context.Context, a application.Avatar) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize replacements including the first insert; keys are generated once and never reused.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, a.Kind+":"+a.OwnerID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO avatar_garbage(object_key,ready_at) SELECT object_key,now() FROM avatars WHERE owner_kind=$1 AND owner_id=$2 AND object_key<>$3 ON CONFLICT(object_key) DO UPDATE SET ready_at=now()`, a.Kind, a.OwnerID, a.Key); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO avatars(owner_kind,owner_id,object_key,bytes,updated_at) VALUES($1,$2,$3,$4,$5) ON CONFLICT(owner_kind,owner_id) DO UPDATE SET object_key=excluded.object_key,bytes=excluded.bytes,updated_at=excluded.updated_at`, a.Kind, a.OwnerID, a.Key, a.Bytes, a.UpdatedAt); err != nil {
		return translate(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM avatar_garbage WHERE object_key=$1`, a.Key); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *Repository) Avatar(ctx context.Context, kind, owner string) (application.Avatar, error) {
	a := application.Avatar{Kind: kind, OwnerID: owner}
	err := r.pool.QueryRow(ctx, `SELECT a.object_key,a.bytes,a.updated_at FROM avatars a WHERE a.owner_kind=$1 AND a.owner_id=$2 AND (($1='user' AND EXISTS(SELECT 1 FROM users u WHERE u.id=a.owner_id AND NOT u.is_blocked)) OR ($1='team' AND EXISTS(SELECT 1 FROM teams t WHERE t.id=a.owner_id AND t.is_approved)))`, kind, owner).Scan(&a.Key, &a.Bytes, &a.UpdatedAt)
	return a, translate(err)
}

// The session lock conflicts with deploy/backup-guard.sh. Keep this exact
// connection checked out until every physical deletion and receipt update ends.
func (r *Repository) TryAvatarCleanupGuard(ctx context.Context) (func(), bool, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return nil, false, err
	}
	var acquired bool
	if err = conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(482161,1)`).Scan(&acquired); err != nil {
		// A lost reply may follow successful acquisition; never pool that session.
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		raw := conn.Hijack()
		_ = raw.Close(cleanup)
		return nil, false, err
	}
	if !acquired {
		conn.Release()
		return nil, false, nil
	}
	release := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var unlocked bool
		if err := conn.QueryRow(cleanup, `SELECT pg_advisory_unlock(482161,1)`).Scan(&unlocked); err != nil || !unlocked {
			// A pooled session must never retain an unaccounted advisory lock.
			raw := conn.Hijack()
			_ = raw.Close(cleanup)
			return
		}
		conn.Release()
	}
	return release, true, nil
}
