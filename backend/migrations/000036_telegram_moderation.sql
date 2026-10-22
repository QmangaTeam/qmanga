-- A separate delivery queue never consumes the in-app notification outbox.
CREATE TABLE moderation_telegram_queue (
 id bigserial PRIMARY KEY,
 kind text NOT NULL,
 entity_id text NOT NULL,
 label text NOT NULL,
 detail text NOT NULL DEFAULT '',
 path text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 delivered_at timestamptz,
 attempts int NOT NULL DEFAULT 0,
 available_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX moderation_telegram_due ON moderation_telegram_queue(available_at,id) WHERE delivered_at IS NULL;

CREATE FUNCTION enqueue_moderation_telegram() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE row_data jsonb:=to_jsonb(NEW); old_data jsonb; item_label text; item_detail text; item_path text; item_kind text;
BEGIN
 IF TG_OP='UPDATE' THEN old_data:=to_jsonb(OLD); END IF;
 IF TG_TABLE_NAME='teams' THEN
  IF NEW.is_approved OR NEW.review_status<>'pending' THEN RETURN NEW; END IF;
  IF TG_OP='UPDATE' AND old_data->>'review_status'='pending' AND (old_data-'updated_at')=(row_data-'updated_at') THEN RETURN NEW; END IF;
  item_kind:='team';item_label:=NEW.name;item_detail:=NEW.description;item_path:='/admin/teams';
 ELSE
  IF row_data->>'status'<>(CASE WHEN TG_TABLE_NAME='revisions' THEN 'submitted' ELSE 'pending' END) THEN RETURN NEW; END IF;
  IF TG_OP='UPDATE' AND old_data->>'status'=row_data->>'status' THEN RETURN NEW; END IF;
  CASE TG_TABLE_NAME
   WHEN 'revisions' THEN
    SELECT title.name_kk||' · Глава '||chapter.label,team.name INTO item_label,item_detail FROM releases publication JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN titles title ON title.id=chapter.title_id JOIN teams team ON team.id=publication.team_id WHERE publication.id=NEW.release_id;
    item_kind:='chapter';item_path:='/admin/moderation?revision='||NEW.id;
   WHEN 'project_requests' THEN
    SELECT coalesce(title.name_kk,nullif(NEW.name_kk,''),'Новый проект'),team.name||E'\n'||NEW.note INTO item_label,item_detail FROM teams team LEFT JOIN titles title ON title.id=NEW.title_id WHERE team.id=NEW.team_id;
    item_kind:='project';item_path:='/admin/requests';
   WHEN 'title_change_requests' THEN
    SELECT title.name_kk,coalesce(team.name,'')||E'\n'||NEW.reason INTO item_label,item_detail FROM titles title LEFT JOIN teams team ON team.id=NEW.team_id WHERE title.id=NEW.title_id;
    item_kind:='title_change';item_path:='/moderation/title-changes/'||NEW.id;
   WHEN 'reports' THEN
    SELECT coalesce(title.name_kk,'Жалоба'),NEW.category||E'\n'||NEW.reason INTO item_label,item_detail FROM (SELECT 1) fallback LEFT JOIN titles title ON title.id=NEW.title_id;
    item_kind:='report';item_path:='/admin/reports';
  END CASE;
 END IF;
 INSERT INTO moderation_telegram_queue(kind,entity_id,label,detail,path) VALUES(item_kind,NEW.id,coalesce(item_label,NEW.id),coalesce(item_detail,''),item_path);
 RETURN NEW;
END $$;
CREATE TRIGGER telegram_team_submission AFTER INSERT OR UPDATE ON teams FOR EACH ROW EXECUTE FUNCTION enqueue_moderation_telegram();
CREATE TRIGGER telegram_revision_submission AFTER INSERT OR UPDATE OF status ON revisions FOR EACH ROW EXECUTE FUNCTION enqueue_moderation_telegram();
CREATE TRIGGER telegram_project_submission AFTER INSERT OR UPDATE OF status ON project_requests FOR EACH ROW EXECUTE FUNCTION enqueue_moderation_telegram();
CREATE TRIGGER telegram_title_submission AFTER INSERT OR UPDATE OF status ON title_change_requests FOR EACH ROW EXECUTE FUNCTION enqueue_moderation_telegram();
CREATE TRIGGER telegram_report_submission AFTER INSERT OR UPDATE OF status ON reports FOR EACH ROW EXECUTE FUNCTION enqueue_moderation_telegram();
-- Existing outstanding items are included once when this migration is applied.
INSERT INTO moderation_telegram_queue(kind,entity_id,label,detail,path)
 SELECT 'team',id,name,description,'/admin/teams' FROM teams WHERE NOT is_approved AND review_status='pending'
 UNION ALL SELECT 'chapter',revision.id,title.name_kk||' · Глава '||chapter.label,team.name,'/admin/moderation?revision='||revision.id FROM revisions revision JOIN releases publication ON publication.id=revision.release_id JOIN chapters chapter ON chapter.id=publication.chapter_id JOIN titles title ON title.id=chapter.title_id JOIN teams team ON team.id=publication.team_id WHERE revision.status='submitted'
 UNION ALL SELECT 'project',request.id,coalesce(title.name_kk,nullif(request.name_kk,''),'Новый проект'),team.name||E'\n'||request.note,'/admin/requests' FROM project_requests request JOIN teams team ON team.id=request.team_id LEFT JOIN titles title ON title.id=request.title_id WHERE request.status='pending'
 UNION ALL SELECT 'title_change',request.id,title.name_kk,coalesce(team.name,'')||E'\n'||request.reason,'/moderation/title-changes/'||request.id FROM title_change_requests request JOIN titles title ON title.id=request.title_id LEFT JOIN teams team ON team.id=request.team_id WHERE request.status='pending'
 UNION ALL SELECT 'report',report.id,coalesce(title.name_kk,'Жалоба'),report.category||E'\n'||report.reason,'/admin/reports' FROM reports report LEFT JOIN titles title ON title.id=report.title_id WHERE report.status='pending';
