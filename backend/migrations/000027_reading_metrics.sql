-- Metrics begin at this migration: historical confirmation times are unknown.
CREATE TABLE metrics_collection (singleton boolean PRIMARY KEY DEFAULT true CHECK(singleton), started_at timestamptz NOT NULL DEFAULT now(), pruned_on date NOT NULL DEFAULT (now() AT TIME ZONE 'UTC')::date);
INSERT INTO metrics_collection(singleton) VALUES(true);
CREATE TABLE metrics_daily(day date NOT NULL,team_id varchar(36) NOT NULL DEFAULT '',metric varchar(32) NOT NULL,value bigint NOT NULL CHECK(value>=0),PRIMARY KEY(team_id,day,metric));
CREATE INDEX metrics_daily_period ON metrics_daily(day,team_id);
CREATE TABLE metrics_publications(release_id varchar(36) PRIMARY KEY REFERENCES releases(id) ON DELETE CASCADE);
-- No historical publication event is manufactured on republishing an old release.
INSERT INTO metrics_publications SELECT id FROM releases WHERE status='published';
CREATE TABLE reading_completions(user_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,title_id varchar(36) NOT NULL REFERENCES titles(id) ON DELETE CASCADE,chapter_id varchar(36) NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,PRIMARY KEY(user_id,title_id,chapter_id));
CREATE TABLE reading_activity(day date NOT NULL,user_id varchar(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,release_id varchar(36) NOT NULL REFERENCES releases(id) ON DELETE CASCADE,completed boolean NOT NULL DEFAULT false,PRIMARY KEY(day,user_id,release_id));
CREATE INDEX reading_activity_release_day ON reading_activity(release_id,day);
CREATE INDEX releases_metrics_team ON releases(team_id,id);
CREATE OR REPLACE FUNCTION metrics_increment(team varchar,kind varchar) RETURNS void LANGUAGE sql AS $$
 INSERT INTO metrics_daily(day,team_id,metric,value) VALUES((now() AT TIME ZONE 'UTC')::date,team,kind,1) ON CONFLICT(team_id,day,metric) DO UPDATE SET value=metrics_daily.value+1;
$$;
CREATE FUNCTION metrics_verified() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF NEW.is_verified AND (TG_OP='INSERT' OR NOT OLD.is_verified) THEN PERFORM metrics_increment('','verified'); END IF; RETURN NEW; END $$;
CREATE TRIGGER metrics_verified AFTER INSERT OR UPDATE OF is_verified ON users FOR EACH ROW EXECUTE FUNCTION metrics_verified();
CREATE FUNCTION metrics_published() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF NEW.status='published' THEN
 INSERT INTO metrics_publications VALUES(NEW.id) ON CONFLICT DO NOTHING;
 IF FOUND THEN PERFORM metrics_increment('','publications'); PERFORM metrics_increment(NEW.team_id,'publications'); END IF;
 END IF; RETURN NEW; END $$;
CREATE TRIGGER metrics_published AFTER INSERT OR UPDATE OF status ON releases FOR EACH ROW EXECUTE FUNCTION metrics_published();
CREATE FUNCTION metrics_rating() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM metrics_increment('','ratings'); RETURN NEW; END $$;
CREATE TRIGGER metrics_rating AFTER INSERT ON ratings FOR EACH ROW EXECUTE FUNCTION metrics_rating();
CREATE FUNCTION metrics_library() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF NOT NEW.is_removed AND (TG_OP='INSERT' OR OLD.is_removed) THEN
 PERFORM metrics_increment('','libraries'); IF NEW.preferred_team_id IS NOT NULL THEN PERFORM metrics_increment(NEW.preferred_team_id,'libraries'); END IF;
 END IF;
 IF NOT NEW.is_removed AND NEW.notify_new_releases AND (TG_OP='INSERT' OR OLD.is_removed OR NOT OLD.notify_new_releases) THEN
 PERFORM metrics_increment('','subscriptions'); IF NEW.preferred_team_id IS NOT NULL THEN PERFORM metrics_increment(NEW.preferred_team_id,'subscriptions'); END IF;
 END IF; RETURN NEW; END $$;
CREATE TRIGGER metrics_library AFTER INSERT OR UPDATE ON library_entries FOR EACH ROW EXECUTE FUNCTION metrics_library();
-- At most once a UTC day; no IP, email, user agent or device fingerprint is stored.
CREATE FUNCTION metrics_prune() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
 IF EXISTS(SELECT 1 FROM metrics_collection WHERE pruned_on<(now() AT TIME ZONE 'UTC')::date) AND pg_try_advisory_xact_lock(814827) THEN
 UPDATE metrics_collection SET pruned_on=(now() AT TIME ZONE 'UTC')::date WHERE pruned_on<(now() AT TIME ZONE 'UTC')::date;
 IF FOUND THEN DELETE FROM reading_activity WHERE day<(now() AT TIME ZONE 'UTC')::date-89; DELETE FROM metrics_daily WHERE day<(now() AT TIME ZONE 'UTC')::date-89; END IF;
 END IF; RETURN NEW; END $$;
CREATE TRIGGER metrics_prune AFTER INSERT ON reading_activity FOR EACH ROW EXECUTE FUNCTION metrics_prune();
