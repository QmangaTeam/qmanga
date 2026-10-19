# Guide продолжения Qmanga и production-контракт

Этот файл — инструкция следующему агенту, а не разрешение менять production без запроса владельца. Обязательные правила также в [AGENTS.md](../../AGENTS.md), полный контракт и восстановление — в [RUNBOOK.md](RUNBOOK.md). Значения секретов здесь не хранить.

## Проверенная точка передачи

2026-10-19: staging и production выпущены на `18eb4fdf02ee050b2cc2915195e03580a4a80594`; внешние HTTPS health/homepage и Google enabled проверены. [Verify Qmanga 39186717189](https://github.com/maqsatto/qmanga/actions/runs/39186717189): acceptance и все три ARM64 images success. Локально 50/50 browser, PostgreSQL/HTTP/codecs, frontend и независимый review. Перед продолжением заново прочитай live state: эта запись историческая.

2026-10-19 (вечер): staging и production выпущены на `f4d73ef9a047810e766f10be930e1089b2e01085` (TASK-62, TASK-63, служебный плагин graphify). [Verify Qmanga 39202510720](https://github.com/maqsatto/qmanga/actions/runs/39202510720): acceptance и все три ARM64 images success; SSH-deploy job штатно skipped. Staging поднят poller-ом, production переведён `release.sh production` на green, предыдущая раскладка (blue) остаётся тёплой для отката. Внешние проверки: health 200, главная отдаёт SHA, `/brand/qmanga.png` и `/brand/qmanga-light.png` 200 на staging и production. Ждёт подтверждения владельца: авто-связка существующих email-аккаунтов при реальном входе Google (TASK-62, статус Review).

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
| Проверенные host keys | `/home/maqsatto/.ssh/qmanga_known_hosts` (0644) |
| Каталог ключей | `/home/maqsatto/.ssh` (0700) |

Ключи перенесены из Downloads; оригиналы удалены по просьбе владельца. Не ищи их снова в Downloads, не выводи содержимое, не коммить и не отключай проверку host key. Эти пути относятся к компьютеру владельца; в другом окружении нужен доступ к тем же защищённым файлам.

```sh
ssh -o BatchMode=yes -o IdentitiesOnly=yes \
  -o StrictHostKeyChecking=yes \
  -o UserKnownHostsFile=/home/maqsatto/.ssh/qmanga_known_hosts \
  -i /home/maqsatto/.ssh/qmanga_deploy qmanga@138.3.254.185
```

Deploy-доступ проверен. Не открывать SSH всему интернету, не создавать Bastion и не менять OCI/firewall ради обычного выпуска. Docker-доступ на сервере даёт фактически root-права.

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

## CI → staging → production

Push `main` запускает Verify Qmanga. Только после acceptance публикуются `ghcr.io/maqsatto/qmanga-{api,web,gateway}:<40sha>` для linux/arm64. Проверить успешный run именно нужного SHA и все три publication jobs. Отдельный GitHub SSH deploy intentionally skipped: используется исходящий poller VM, обычно каждые5минут. Не включать второй способ деплоя.

После CI staging обновляется автоматически. При необходимости штатный ручной выпуск на VM:

```sh
bash /opt/qmanga/repo/deploy/release.sh staging <tested-40-character-sha>
```

Проверить `staging.active`, локальные API и главную по активному порту с заголовком `Host: qmanga-dev.maqsatto.dev`; при необходимости открыть сценарии через SSH-туннель. Публичного dev-сайта и маршрута Caddy больше нет. Затем **тот же SHA**:

```sh
bash /opt/qmanga/repo/deploy/release.sh production <same-tested-40-character-sha>
```

Скрипт сам проверяет ancestry main, staging SHA, restore gate, делает свежий backup с проверкой вне VM, поднимает второй цвет, переключает Caddy, проверяет здоровье и оставляет прошлый stack для rollback. Не обходить lock/gates и не заменять это ручным `compose up`. Если deployment lock занят, дождаться текущего запуска и перечитать state.

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
