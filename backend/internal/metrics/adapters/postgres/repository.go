package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qmanga/internal/metrics/application"
	"time"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }
func (repository *Repository) Report(ctx context.Context, actor string, admin bool, days int) (application.Report, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	result := application.Report{Days: []application.Day{}}
	tx, err := repository.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users u WHERE u.id=$1 AND u.is_verified AND NOT u.is_blocked AND (($2 AND u.role='admin') OR (NOT $2 AND EXISTS(SELECT 1 FROM team_memberships m JOIN teams t ON t.id=m.team_id AND t.is_approved WHERE m.user_id=u.id))))`, actor, admin).Scan(&allowed)
	if err != nil {
		return result, err
	}
	if !allowed {
		return result, application.ErrForbidden
	}
	if err = tx.QueryRow(ctx, `SELECT started_at FROM metrics_collection`).Scan(&result.StartedAt); err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `WITH days AS (SELECT generate_series((now() AT TIME ZONE 'UTC')::date-($3::int-1),(now() AT TIME ZONE 'UTC')::date,'1 day')::date AS day), scoped AS MATERIALIZED (SELECT d.day,d.metric,sum(d.value) value FROM metrics_daily d WHERE d.day>=(now() AT TIME ZONE 'UTC')::date-($3::int-1) AND d.day<=(now() AT TIME ZONE 'UTC')::date AND (($2 AND d.team_id='') OR (NOT $2 AND d.team_id IN(SELECT m.team_id FROM team_memberships m JOIN teams t ON t.id=m.team_id AND t.is_approved WHERE m.user_id=$1))) GROUP BY d.day,d.metric), activity AS MATERIALIZED (SELECT a.day,count(DISTINCT a.user_id) readers,count(*) FILTER(WHERE a.completed) completions FROM reading_activity a JOIN releases r ON r.id=a.release_id WHERE a.day>=(now() AT TIME ZONE 'UTC')::date-($3::int-1) AND a.day<=(now() AT TIME ZONE 'UTC')::date AND ($2 OR r.team_id IN(SELECT m.team_id FROM team_memberships m JOIN teams t ON t.id=m.team_id AND t.is_approved WHERE m.user_id=$1)) GROUP BY a.day) SELECT to_char(days.day,'YYYY-MM-DD'),coalesce(sum(s.value) FILTER(WHERE s.metric='verified'),0)::bigint,coalesce(sum(s.value) FILTER(WHERE s.metric='publications'),0)::bigint,coalesce(a.readers,0),coalesce(a.completions,0),coalesce(sum(s.value) FILTER(WHERE s.metric='ratings'),0)::bigint,coalesce(sum(s.value) FILTER(WHERE s.metric='libraries'),0)::bigint,coalesce(sum(s.value) FILTER(WHERE s.metric='subscriptions'),0)::bigint FROM days LEFT JOIN scoped s ON s.day=days.day LEFT JOIN activity a ON a.day=days.day GROUP BY days.day,a.readers,a.completions ORDER BY days.day`, actor, admin, days)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var day application.Day
		if err = rows.Scan(&day.Date, &day.Verified, &day.Publications, &day.Readers, &day.Completions, &day.Ratings, &day.Libraries, &day.Subscriptions); err != nil {
			rows.Close()
			return result, err
		}
		result.Days = append(result.Days, day)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	err = tx.QueryRow(ctx, `SELECT count(DISTINCT a.user_id) FROM reading_activity a JOIN releases r ON r.id=a.release_id WHERE a.day>=(now() AT TIME ZONE 'UTC')::date-($3::int-1) AND a.day<=(now() AT TIME ZONE 'UTC')::date AND ($2 OR r.team_id IN(SELECT m.team_id FROM team_memberships m JOIN teams t ON t.id=m.team_id AND t.is_approved WHERE m.user_id=$1))`, actor, admin, days).Scan(&result.PeriodReaders)
	if err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}
