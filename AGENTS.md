# Работа над Qmanga

- Делай то, что просит владелец. Никаких карточек, планов и отчётов на каждую задачу.
- Архитектура: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md). Интерфейс: `apps/web/AGENTS.md` и [docs/ui/DESIGN-REVIEW.md](docs/ui/DESIGN-REVIEW.md).
- Небольшие логичные коммиты прямо в `main`; push — по просьбе владельца.
- Проверка: `go -C backend build ./... && go -C backend vet ./...` и `bun --cwd apps/web check`. Unit-тесты — по необходимости, CI их не запускает.
- Production: push в `main` → Fast Release собирает образы и двигает `refs/deploy-ready` → poller на VM выкатывает staging, затем production. Перед серверной работой прочитай [OPERATIONS](docs/deployment/OPERATIONS.md) и [RUNBOOK](docs/deployment/RUNBOOK.md). Не печатай секреты, не удаляй production volumes, не правь VM вручную, не обходи backup/lock.
- Миграции должны быть совместимы с предыдущим релизом (окно отката).
