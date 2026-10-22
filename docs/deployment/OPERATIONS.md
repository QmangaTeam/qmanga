# Guide продолжения Qmanga и production-контракт

Этот файл — инструкция следующему агенту, а не разрешение менять production без запроса владельца. Обязательные правила также в [AGENTS.md](../../AGENTS.md), полный контракт и восстановление — в [RUNBOOK.md](RUNBOOK.md). Значения секретов здесь не хранить.

## Проверенная точка передачи

2026-10-19: staging и production выпущены на `18eb4fdf02ee050b2cc2915195e03580a4a80594`; внешние HTTPS health/homepage и Google enabled проверены. [Verify Qmanga 39186717189](https://github.com/maqsatto/qmanga/actions/runs/39186717189): acceptance и все три ARM64 images success. Локально 50/50 browser, PostgreSQL/HTTP/codecs, frontend и независимый review. Перед продолжением заново прочитай live state: эта запись историческая.

2026-10-19 (вечер): staging и production выпущены на `f4d73ef9a047810e766f10be930e1089b2e01085` (TASK-62, TASK-63, служебный плагин graphify). [Verify Qmanga 39202510720](https://github.com/maqsatto/qmanga/actions/runs/39202510720): acceptance и все три ARM64 images success; SSH-deploy job штатно skipped. Staging поднят poller-ом, production переведён `release.sh production` на green, предыдущая раскладка (blue) остаётся тёплой для отката. Внешние проверки: health 200, главная отдаёт SHA, `/brand/qmanga.png` и `/brand/qmanga-light.png` 200 на staging и production. Ждёт подтверждения владельца: авто-связка существующих email-аккаунтов при реальном входе Google (TASK-62, статус Review).

2026-10-20: staging и production выпущены на `bb8aec38d6c019e08d940b6209e794795ca956e3` (TASK-64: 719 меток Senkuro; TASK-65: панель фильтров каталога). [Verify Qmanga 39333412292](https://github.com/maqsatto/qmanga/actions/runs/39333412292): обе acceptance-половины и все три ARM64 images success. Staging poller поднял тот же SHA (green), production переведён `release.sh production` на blue, прошлый green (`f4d73ef`) остаётся тёплым для отката; backup `20261019T222451Z` проверен off-host. Внешние проверки: health 200, главная отдаёт SHA, `/api/v1/genres` = 713, 6 групп в порядке main-cast→demographics, панель `/catalog` проверена Playwright (kk/ru, обе темы, desktop/mobile).

2026-10-20 (ночь): включён автоматический путь `push → staging → production` (TASK-66) и проверен серией реальных релизов. Текущая production-точка: `36b3a5f455eb37830408af44c2e2f00b48eabd51` (TASK-67: desktop-панель фильтров каталога снова скроллится). Ей предшествовали автоматические выпуски: `a06546e` (исправления независимого review: тег компонента по последнему изменению входов, fail-closed при отсутствующем образе, сверка запущенных image ID, отказ от re-exec; slow path 1012 s), `71d4d2b` (снятие меток приёмки, web+api, 332 s), double-push `fa5152d`→`c299a2d` (315 s, production на втором SHA), slow path `5c98ee8` (685 s). Замеры push → публичный production при старом 5-минутном таймере: CSS 350 s, Svelte 300 s, Go 351 s, web+api 332 s; сборка 34–102 s. Drill `504d3ab`: намеренный сбой health в контейнере — production остался на прежнем SHA, плохой SHA записан в `state/deployment-hold`. Verify для `36b3a5f` и `a06546e` зелёный; docs-only push `5dacbbd` не создал workflow-запусков и не сдвинул `refs/deploy-ready`. Осталось владельцу: установить обновлённые systemd service/timer (команда ниже), после чего ожидаемое время ~130–260 s.

До этого автоматического пути production последовательно был на `bb8aec3`, `43c1a5d`, `cf45ae7`, `8d7d2dd`, а staging выпускался тем же poller-ом.

2026-10-20 (18:42 UTC): staging и production автоматически выпущены штатным poller-ом на `e2c9c8910d6dedbe74eddbafa028b8bb67180340` (CBZ/PDF, Shikimori/AniList и автоматические теги, исправление публикации canonical component tags). [Verify Qmanga 39474894775](https://github.com/maqsatto/qmanga/actions/runs/39474894775) и [Fast Release 39474894489](https://github.com/maqsatto/qmanga/actions/runs/39474894489) success; обе acceptance-половины, 88 браузерных сценариев (42+46) и три ARM64 image publication прошли. Независимый review исправления CI завершён после устранения гонок HEAD/alias. Staging backup `20261020T183640Z`, production backup `20261020T184220Z`, последняя копия проверена off-host перед миграцией. Оба active-файла — green того же SHA, image ID проверены release.sh; предыдущий blue сохранён. Внешние HTTPS health/homepage/genres/PDF assets — 200, новый private metadata-import route — 401 без входа, SHA подтверждён в браузере. Приватный production сценарий импорта под аккаунтом владельца и физический Android остаются непроверенными; импортные карточки сохраняют Review. Старые 5-минутные systemd units действуют до установки владельцем, поэтому ≤300s пока не подтверждено.

2026-10-21 00:15 (Asia/Aqtau; 20 октября 19:15 UTC): staging и production автоматически выпущены на `797c810725aa6e880cabd477a8a2a2570e4b49a0` (TASK-86: фокус при открытии/отмене inline-редактора аккаунта). [Verify39480229990](https://github.com/maqsatto/qmanga/actions/runs/39480229990) и [FastRelease39480229999](https://github.com/maqsatto/qmanga/actions/runs/39480229999) success; 89 browser cases (43+46), 0 failures. Независимое review и повторный реальный Docker account-search2/2 pass. На фактическом staging того же SHA выполнены: CBZ/PDF ordered preview→cancel без загрузок/страниц; Shikimori search/select/prefill; AniList search/select/canonical tags/WebP; guest401/reader403/invalid-source422; account edit→focus→cancel→origin focus. Только временные синтетические staging-сущности, после проверки удалены с подтверждением0 оставшихся аккаунтов/title/team; временные credentials удалены. Внешний production health/PostgreSQL/S3, homepage SHA и PDF assets подтверждены, предыдущий green сохранён. UX TASK69–98 (кроме физического TASK97) и импорты99–103 приняты и закрыты. Требование отдельного входа владельца в production сверх AC снято; физический Android и установка20секундных units с sudo остаются внешними действиями.

2026-10-21 10:00 (Asia/Aqtau; 05:00 UTC): staging и production автоматически выпущены poller-ом на `1cdb00a11d9ef8a24c296a9a117501f44ebae203` (TASK-104/106/107: responsive title metadata, single-Q branding, auth/mobile navigation, comments UI). [Fast Release39546480513](https://github.com/maqsatto/qmanga/actions/runs/39546480513) и [Verify39546480550](https://github.com/maqsatto/qmanga/actions/runs/39546480550) success; обе acceptance-половины,98 browser cases (45+53),0fail0skip. Исправлены cold-host конфигурация/cache standalone Vite browser tests и accessible label статистики; независимый финальный review без блокеров. Оба active-файла blue того же SHA; прежний green сохранён для rollback. Внешний HTTPS health/PostgreSQL/S3 и SHA подтверждены; public Playwright kk/ru × dark/light,320/360/1280 проверил branding/login/register/mobile nav/title metadata/tabs/description/nooverflow. Backend и миграции не менялись, локальные CBZ/fixture accounts не переносились. Физический Android остаётся вне этой проверки.

2026-10-22 15:54 (Asia/Aqtau;10:54UTC): poller автоматически выпустил staging и production на green `4e4b8d87ac01005883f5d4656992a4fe177aeaa6` (накопленный frontend-пакет: главная, тайтл/комментарии, HTML/Figma админка и студия, обзор проекта и редактор). [Fast Release39765129545](https://github.com/maqsatto/qmanga/actions/runs/39765129545) success; оба active-файла и публичная главная подтверждают тот же SHA, HTTPS health PostgreSQL/S3/status ok. Read-only браузерная проверка: staging36 и production48 responsive views,kk/ru,dark/light,320/360/1280,0runtime exceptions, гостевой вход в Studio через sign-in. Backend/миграции и production-данные не менялись. Полный Verify пока не подтверждён: старые UI-тесты ожидали удалённые/свёрнутые блоки; test-only исправления опубликованы вплоть до `99fd185a6bd67a012ddb06b22fd53536a688c23e`, локально project-cards1/1,form-workflow3/3,metrics2/2 pass и независимый review без замечаний. [Verify39766638560](https://github.com/maqsatto/qmanga/actions/runs/39766638560) нужно проверить по фактическому результату; его успех на момент этой записи не заявлен. Новый test-only SHA также идёт штатной автоматической цепочкой; текущий production перед следующей работой сверять заново. На VM всё ещё пятиминутный poller; staging и production занимают отдельные тики.

## Адреса и доступ

| Назначение | Адрес / путь |
| --- | --- |
| Репозиторий | https://github.com/maqsatto/qmanga |
| Локальный checkout владельца | `/home/maqsatto/Documents/projects/qmanga` |
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

1. Прочитать root/web AGENTS и нужную карточку через `bun run backlog task view TASK-ID --plain`; менять задачи только CLI. Сохранять дословные формулировки владельца отдельно от пояснения агента.
2. Реализовать в локальном checkout, соблюдая Clean Architecture; миграции добавлять вперёд, совместимыми с предыдущим production API. Не использовать реальные данные владельца как fixtures. Не менять роли/почту/пароль повторно из-за исторической записи.
3. Проверить затронутые границы по RUNBOOK: реальный PostgreSQL/codecs для persistence/media, браузерные взаимодействия для UI,320/360/desktop,kk/ru,темы. Для полного Go: `sudo docker compose --profile tools run --rm go-tests go test ./... -count=1`. Для web: `bun run --cwd apps/web check`, build и профильные тесты. Docker локально может требовать sudo.
4. Полный fixture browser suite требует игнорируемый `QMANGA_FIXTURE_FILE`; не подставлять prod. CI запускает файлы последовательно группами по8 с recreate того же API между группами, чтобы не исчерпывать реальный auth limit30/min. Не ослаблять лимит продукта ради тестов; не перезапускать сервис во время активной группы.
5. Проверить diff, сделать небольшой коммит; push только при авторизации владельца. В этой сессии выпуск был явно запрошен, это не постоянное разрешение всем будущим задачам.
6. Для нетривиального завершения запросить независимый review субагентом и исправить существенные находки. Review подтверждает код, а не факт выпуска.

## CI → staging → production (автоматический быстрый путь)

С 2026-10-20 обычный `git push origin main` доходит до production без ручных команд. Работают два независимых workflow:

1. **Fast Release** (`.github/workflows/release.yml`, блокирующий). Сравнивает файлы с последним выпущенным SHA (`refs/deploy-ready`), собирает только изменённые компоненты (`web` — `apps/web/`; `api`+worker — `backend/`; `gateway` — nginx/Caddy) с GHA Docker cache, пушит immutable arm64 теги SHA workflow и alias последнего изменения входов компонента и двигает `refs/deploy-ready`. Push только с документацией ничего не выпускает. Изменения `backend/migrations/`, `deploy/`, `infra/`, `.github/` — медленный путь: workflow ждёт успешный `Verify Qmanga` этого же SHA и только затем двигает ref.
2. **Full Verification** (`verify.yml`) — полный Go/PostgreSQL/codec/browser acceptance и публикация всех трёх arm64 images. Он больше не блокирует обычный релиз; его результат виден в Actions.

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
