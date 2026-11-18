# Guide продолжения Qmanga и production-контракт

Этот файл — инструкция следующему агенту, а не разрешение менять production без запроса владельца. Обязательные правила также в [AGENTS.md](../../AGENTS.md), полный контракт и восстановление — в [RUNBOOK.md](RUNBOOK.md). Значения секретов здесь не хранить.

## Проверенная точка передачи

Старые записи о выкатах удалены; актуальный SHA смотри на сервере (`/opt/qmanga/state/production.active`) и на главной странице.

2026-10-23 (Asia/Aqtau): рабочий репозиторий перенесён в [QmangaTeam/qmanga](https://github.com/QmangaTeam/qmanga), организация называется Qmanga Team. `maqsatto` и `Bauka07` — Owner, базовая роль участников — Read. [Verify 39844622844](https://github.com/QmangaTeam/qmanga/actions/runs/39844622844) и [Fast Release 39844622864](https://github.com/QmangaTeam/qmanga/actions/runs/39844622864) success для `90b6db03f59dfaf1a19daa40e1cb7ad0d4bf7c6b`; poller автоматически выпустил оба окружения на blue того же SHA, предыдущий green — `1722e837c4d3ad9011cf9479e854c10cd45ba30c`. Внешние HTTPS health PostgreSQL/S3 и SHA главной подтверждены. [Миграция registry 39843716253](https://github.com/QmangaTeam/qmanga/actions/runs/39843716253) скопировала current/previous immutable tags с проверкой manifest SHA256; на Oracle ARM64 все пять refs реально загружены без credentials, image IDs совпали с исходными. Все четыре production-сервиса используют `ghcr.io/qmangateam/qmanga-*`; staging этого переходного выпуска сохраняет прежние имена `ghcr.io/maqsatto/qmanga-*`, но фактические image IDs всех четырёх сервисов совпадают с перенесёнными. Следующий обычный staging-релиз прочитает новый default; принудительный перезапуск ради имён не нужен. Оба host env содержат новый IMAGE_PREFIX с сохранением прав `0600`, timer активен, hold для текущего SHA отсутствует. Previous manifests доступны в обоих namespaces с одинаковыми IDs; старые packages сохранять для перехода и отката, реальный rollback в этой приёмке не запускался. Три новых packages публичны; временное разрешение public package creation после настройки возвращено в false. Локальный origin — SSH существующим ключом, VM origin — новый публичный HTTPS URL; новые SSH/Oracle-права Bauka не выдавались по уточнению владельца.

2026-10-23: owner-authorized public UI follow-ups deployed through the normal poller to production `5039821661afe5eb4b9711098a7bc653d3a63595`. [Fast Release39941710159](https://github.com/QmangaTeam/qmanga/actions/runs/39941710159) succeeded and deploy-ready/public footer match. External health200 (PostgreSQL/S3 ok); live ranking, Load more, catalogue sorting and title Comments transitions checked. Approximate push-to-public interval10minutes; no VM changes and no local fixtures deployed. [Full Verify39941710199](https://github.com/QmangaTeam/qmanga/actions/runs/39941710199) failed both browser integration shards; full CI is not green. Exact assertions remain to inspect with available authenticated logs. SSH port22 currently times out despite existing documented key; do not change firewall or deployment gates to bypass this. Authenticated production library interaction and native Figma acceptance remain unverified.



## Адреса и доступ

| Назначение | Адрес / путь |
| --- | --- |
| Репозиторий | https://github.com/QmangaTeam/qmanga |
| Локальный checkout владельца | `/home/maqsatto/qmanga` |
| Production | https://qmanga.maqsatto.dev |
| Staging | Только на VM: `127.0.0.1:18280` (blue) или `127.0.0.1:18281` (green); цвет в `staging.active` |
| Oracle ARM64 VM | `138.3.254.185` |
| Deploy SSH user | `qmanga` |
| Основной приватный ключ | `/home/maqsatto/.ssh/qmanga_deploy` (0600) |
| Дополнительный Oracle ключ | `/home/maqsatto/.ssh/qmanga_oracle` (0600); не подменять им проверенный deploy-доступ |
| Проверенные host keys | `/home/maqsatto/.ssh/known_hosts` (0644) |
| Каталог ключей | `/home/maqsatto/.ssh` (0700) |

Ключи перенесены из Downloads; оригиналы удалены по просьбе владельца. Не ищи их снова в Downloads, не выводи содержимое, не коммить и не отключай проверку host key. Эти пути относятся к компьютеру владельца; в другом окружении нужен доступ к тем же защищённым файлам.

```sh
ssh -o BatchMode=yes -o IdentitiesOnly=yes \
  -o StrictHostKeyChecking=yes \
  -o UserKnownHostsFile=/home/maqsatto/.ssh/known_hosts \
  -i /home/maqsatto/.ssh/qmanga_deploy qmanga@138.3.254.185
```

На 2026-10-20 ключ `qmanga_deploy` присутствует с правами `0600`, проверенные записи VM находятся в `known_hosts`; отдельного `qmanga_known_hosts` нет. После явного согласия владельца единственное правило OCI для TCP 22 обновлено с устаревшего операторского `/32` на текущий `/32`; правило сохранено и проверено в интерфейсе. SSH восстановлен с прежними ключом и проверкой host key. Не открывать SSH всему интернету, не создавать Bastion и не менять OCI/firewall ради обычного выпуска. У deploy-пользователя нет беспарольного sudo, поэтому установка обновлённых systemd service/timer с интервалом 20 секунд остаётся действием владельца. Docker-доступ на сервере даёт фактически root-права.

## Что находится на сервере

- Checkout: `/opt/qmanga/repo`; штатно detached HEAD на выпущенном SHA. Не редактировать продукт вручную на сервере.
- Секреты: `/opt/qmanga/config/staging.env`, `/opt/qmanga/config/production.env`, оба0600. Не печатать их и не переносить в Git. Образец переменных: `deploy/host.env.example`.
- Текущий цвет/SHA: `/opt/qmanga/state/staging.active`, `production.active`; предыдущие: `*.previous`.
- Restore gate: `/opt/qmanga/state/production.restore-verified`; это результат настоящего restore drill, не файл для обхода проверки.
- Данные: Compose projects `qmanga-staging-data`, `qmanga-production-data`, отдельные PostgreSQL/MinIO volumes. Приложения `qmanga-{staging,production}-{blue,green}`; общий TLS proxy `qmanga-front`.
- Порты только loopback: production blue18180/green18181; staging blue18280/green18281. Снаружи только HTTPS, HTTP redirect и ограниченный SSH.
- Бэкапы: `/opt/qmanga/backups/{staging,production}`; off-host hook `/opt/qmanga/hooks/offsite-backup`, OCI CLI `/opt/qmanga/bin/oci`, приватный OCI bucket `qmanga-backups`. Используются instance principals; новые постоянные OCI ключи не нужны.
- Таймеры: `qmanga-staging-poll.timer`, `qmanga-production-backup.timer`. Production backup около02:30UTC. Следить за диском/квотой OCI: автоматической retention пока нет.

```sh
cat /opt/qmanga/state/staging.active /opt/qmanga/state/production.active
systemctl status qmanga-staging-poll.timer qmanga-production-backup.timer
journalctl -u qmanga-staging-poll.service -n 60 --no-pager
journalctl -u qmanga-production-backup.service -n 60 --no-pager
docker ps --format 'table {{.Names}}\t{{.Status}}'
```

Не использовать `docker inspect`, `compose config` без `--quiet`, вывод env или полные HTTP cookies в общие логи: они могут раскрыть секреты.

## Как добавлять изменения

1. Прочитать root/web AGENTS; менять только нужное для задачи.
2. Реализовать в локальном checkout, соблюдая Clean Architecture; миграции добавлять вперёд, совместимыми с предыдущим production API. Не использовать реальные данные владельца как fixtures. Не менять роли/почту/пароль повторно из-за исторической записи.
3. Проверить затронутые границы по RUNBOOK: реальный PostgreSQL/codecs для persistence/media, браузерные взаимодействия для UI,320/360/desktop,kk/ru,темы. Для полного Go: `sudo docker compose --profile tools run --rm go-tests go test ./... -count=1`. Для web: `bun run --cwd apps/web check`, build и профильные тесты. Docker локально может требовать sudo.
4. Полный fixture browser suite требует игнорируемый `QMANGA_FIXTURE_FILE`; не подставлять prod. CI запускает файлы последовательно группами по8 с recreate того же API между группами, чтобы не исчерпывать реальный auth limit30/min. Не ослаблять лимит продукта ради тестов; не перезапускать сервис во время активной группы.
5. Проверить diff, сделать небольшой коммит; push только при авторизации владельца. В этой сессии выпуск был явно запрошен, это не постоянное разрешение всем будущим задачам.
6. Для нетривиального завершения запросить независимый review субагентом и исправить существенные находки. Review подтверждает код, а не факт выпуска.

## CI → staging → production (автоматический быстрый путь)

С 2026-10-20 обычный `git push origin main` доходит до production без ручных команд. Работают два независимых workflow:

1. **Fast Release** (`.github/workflows/release.yml`, блокирующий). Сравнивает файлы с последним выпущенным SHA (`refs/deploy-ready`), собирает только изменённые компоненты (`web` — `apps/web/`; `api`+worker — `backend/`; `gateway` — nginx/Caddy) с GHA Docker cache, пушит immutable arm64 теги SHA workflow и alias последнего изменения входов компонента и двигает `refs/deploy-ready`. Push только с документацией ничего не выпускает. Изменения `backend/migrations/`, `deploy/`, `infra/`, `.github/` — пересобираются все три компонента, затем двигается ref.
2. Отдельного workflow с тестами (`verify.yml`) больше нет.

VM-poller (`qmanga-staging-poll.timer`) читает `refs/deploy-ready` через `git fetch` (без GitHub API/token) и вызывает `release.sh staging`, а на следующем тике — автоматически `release.sh production` того же SHA (poller не делает re-exec: каждая итерация берёт свежий код из репозитория). Интервал в репозитории — 20 с, но на VM до установки файлов ниже действует старый 5-минутный юнит. `release.sh`:

- оба workflow публикации сериализованы по компоненту, сохраняют уже существующий HEAD-тег при повторном запуске и создают alias последнего изменения входов из готового образа workflow HEAD только при отсутствии alias; существующий тег сохраняется, ошибки registry/auth/network останавливают публикацию;
- выбирает тег каждого компонента как последний коммит ≤ SHA, менявший его входы сборки (`apps/web`+`infra/web.Dockerfile`; `backend`+`infra/api.Dockerfile`/`imagemagick-policy.xml`; `infra/nginx.conf`+`gateway.Dockerfile`), и использует только опубликованный immutable-образ. Это корректно, даже если предыдущий релиз упал и `staging.active` отстал от `deploy-ready`; если нужного образа нет, релиз падает с явной ошибкой `component X: image ... is not published` и hold, а не молча подставляет устаревший компонент. CSS-изменение не пересобирает и не перезапускает API/gateway;
- пропускает полный backup, если есть проверенная копия моложе 30 ч (`QMANGA_MAX_BACKUP_AGE_H`, можно задать в systemd-юните poller-а) и нет изменений в `backend/migrations/`; при миграциях, отсутствии/устаревании маркера `COMPLETE` или первом релизе полный backup с off-host hook обязателен; staging и production используют одну политику;
- сохраняет blue-green, deployment lock, staging gate, restore gate, короткие health checks и автоматический возврат маршрута/worker при ошибке; схема БД автоматически не откатывается;
- проверяет, что публичная главная отвечает и отдаёт выпущенный SHA (footer `BUILD_VERSION`), и что контейнеры нового цвета используют именно выбранные теги (`docker inspect` image ID против выбранного immutable-тега), а не кэш прошлого цвета или старый компонент;
- при провале release или ручном `rollback.sh` пишет `state/deployment-hold` с плохим SHA; poller не выкатывает этот SHA повторно, пока `deploy-ready` не уедет на новый. Снять hold вручную: удалить `/opt/qmanga/state/deployment-hold` (только осознанно).

Обновление systemd-юнитов на VM после изменения их файлов (единственная ручная операция для инфраструктуры CI):

```sh
sudo install -m 644 /opt/qmanga/repo/deploy/qmanga-staging-poll.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl restart qmanga-staging-poll.timer
```

Проверка состояния и таймингов:

```sh
cat /opt/qmanga/state/staging.active /opt/qmanga/state/production.active
ls -l /opt/qmanga/state/manifests/
journalctl -u qmanga-staging-poll.service -n 80 --no-pager
curl -fsS https://qmanga.maqsatto.dev/api/v1/health
curl -fsS https://qmanga.maqsatto.dev/ | grep -o <40-sha>
```

Ручной `release.sh` остаётся для согласованных случаев, когда автоматический путь осознанно приостановлен. Не включать второй (SSH) способ деплоя. Fail-fast: SHA, который упал или был откатан, не выкатывается повторно, пока `deploy-ready` не уедет на другой SHA (hold по конкретному SHA, без таймера); исправление выпускается новым коммитом. Если образ нужного компонента не опубликован (например, сборку отменили или удалили пакет), релиз останавливается с сообщением `component X: image ... is not published` — повтори Fast Release/Verify для этого SHA; снимать hold вручную только осознанно.

Одинаковый уже активный SHA — no-op: **не применит новые env**. Секреты/конфигурация требуют согласованного нового выпуска; не считать запись env фактом включения сервиса.

```sh
curl --fail -H 'Host: qmanga-dev.maqsatto.dev' http://127.0.0.1:18280/api/v1/health # blue; для green — 18281
curl --fail https://qmanga.maqsatto.dev/api/v1/health
curl --fail https://qmanga.maqsatto.dev/api/v1/auth/google/status
```

Внутри VM OCI public-IP hairpin не работает: production deployment использует `curl --resolve DOMAIN:443:127.0.0.1`; отдельно проверить production HTTPS извне. Только после наблюдаемого выпуска и требуемой приёмки ставить продуктовой задаче Done; локальный успех означает Review. Записывать SHA/результаты/остаток в ту же карточку, логи и screenshots — ignored tmp.

## Google и аккаунты

Google client pair уже настроен в закрытых env обоих окружений. Callback: `https://<domain>/api/v1/auth/google/callback`; только `openid email profile`. Не менять Cloud redirect/scopes без необходимости. Новый Google аккаунт получает имя/email/avatar из проверенного профиля; повторный вход не затирает ручные изменения. Google с подтверждённой почтой, совпавшей с незаблокированным аккаунтом, зарегистрированным по паролю, автоматически привязывается и входит в этот аккаунт (решение владельца 19.10.2026, TASK-62); заблокированный аккаунт и Google, уже привязанный к другому аккаунту, получают отказ.

Для существующего аккаунта: обычный вход → настройки → «Привязать Google». Нужна подтверждённая почта; flow привязан к точной сессии, повтор/смена/выход отвергаются. Не связывать issuer/sub вручную по email. Настоящую привязку владельца не считать выполненной по mock-тестам. Смена почты требует подтверждения нового адреса; роли/проекты/библиотека сохраняют ID аккаунта. Персональные пароли не хранить в Guide/Backlog.

## Откат и запреты

```sh
bash /opt/qmanga/repo/deploy/rollback.sh production
# либо staging, если проблема только там
```

Откат меняет приложение/маршрут, **не БД и не пользовательские записи**. Не делать schema rollback, `down -v`, удаление volumes, restore поверх живой БД или force-push. Restore — отдельная maintenance-операция с решением владельца; drill только в новых изолированных volumes. Не удалять/перезаписывать media вне существующего backup guard. Один VM остаётся единой точкой отказа.

## Telegram moderation notifications

The existing Docker worker sends moderation notices when both `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` are configured in the private production env file. Leave both blank in staging/local review; no token is passed to API/web. After `/start` in the bot, record the owner's confirmed private chat ID through the secret configuration channel. Never log Bot API URLs, tokens, raw responses or message bodies.

Migration000036 snapshots submissions for chapter revisions, team applications/resubmissions, project requests, title changes and reports into a separate durable queue, including outstanding items once at migration. Delivery retries with exponential delay up to one hour; Telegram failure never changes moderation state. A crash after Telegram accepts a message but before database acknowledgement can cause a duplicate (at-least-once). Messages contain bounded plain text and the exact entity ID; chapters/title changes link directly, while team/project/report buttons explicitly open their corresponding paginated queue. Review and decisions still require site authentication. Queue delivery and pending counts can be inspected without reading message bodies or secrets; worker logs use sanitized errors. Releasing a new SHA recreates the worker and applies changed env; the existing blue/green worker stop/start and rollback lifecycle remains authoritative.
