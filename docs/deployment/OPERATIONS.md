# Qmanga deployment operations

- Push to `main`: CI runs acceptance tests and publishes immutable ARM64 images. The VM checks the result every five minutes and deploys a successful commit to `qmanga-dev.maqsatto.dev` with a health gate and automatic route rollback on failure.
- Promote the commit serving staging to production from the VM as the `qmanga` user, working in `/opt/qmanga/repo`: `bash deploy/release.sh production <40-character staging SHA>`. The release refuses a different SHA, checks the restore marker, makes a verified off-host backup, then switches the route after health checks.
- If the new application must be reverted: `bash deploy/rollback.sh production`. Database migrations and user writes are not rolled back.
- A systemd timer runs `qmanga-production-backup.service` daily at about 02:30 UTC. Check `systemctl status qmanga-production-backup.timer` and `journalctl -u qmanga-production-backup.service`. Each run verifies local checksums, uploads to the private Oracle bucket, downloads every object for comparison and writes `COMPLETE` last.

Both sites use separate PostgreSQL and MinIO volumes on the same VM. The VM remains a single point of failure. Backup storage has no automatic retention policy yet; monitor available disk and bucket capacity as the catalogue grows.
