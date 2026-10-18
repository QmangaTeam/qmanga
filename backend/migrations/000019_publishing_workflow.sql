ALTER TABLE pages ADD COLUMN original_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE revisions ADD COLUMN preparation_method VARCHAR(24) NOT NULL DEFAULT 'unspecified' CHECK (preparation_method IN ('unspecified','manual','machine_assisted','machine'));
ALTER TABLE revisions ADD COLUMN human_reviewed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE revisions ADD COLUMN human_reviewed_by VARCHAR(36) REFERENCES users(id);
ALTER TABLE revisions ADD COLUMN human_reviewed_at TIMESTAMPTZ;
ALTER TABLE titles ADD COLUMN cover_managed BOOLEAN NOT NULL DEFAULT false;
CREATE TABLE title_cover_assets (
 id VARCHAR(36) PRIMARY KEY, title_id VARCHAR(36) NOT NULL REFERENCES titles(id),
 object_key TEXT NOT NULL, mime VARCHAR(32) NOT NULL, checksum VARCHAR(64) NOT NULL,
 width INT NOT NULL, height INT NOT NULL, byte_size BIGINT NOT NULL,
 selected_by VARCHAR(36) REFERENCES users(id), created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Preserve historical derived covers without keeping superseded page URLs public.
INSERT INTO title_cover_assets(id,title_id,object_key,mime,checksum,width,height,byte_size)
SELECT md5(title.id || ':' || variant.id),title.id,variant.object_key,variant.mime,variant.checksum,variant.width,variant.height,variant.byte_size
FROM titles title JOIN page_variants variant ON variant.id=title.cover_asset_id AND variant.tier<>'original';
UPDATE titles title SET cover_asset_id=cover.id FROM title_cover_assets cover WHERE cover.title_id=title.id;
