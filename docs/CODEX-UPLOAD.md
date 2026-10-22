# Codex → Qmanga: короткий гайд

1. Администратор: **Админ-панель → API-токены → Создать**. Включи **«Публиковать без модерации»**, скопируй секрет один раз в приватный файл, например `~/.config/qmanga/upload.token` (`chmod 600`). Не записывай токен в Git, чат или аргументы команды.
2. Подготовь папку JPEG/PNG/WebP или CBZ и `chapter.json`:

```json
{
  "project_id": "ID проекта своей одобренной команды",
  "chapter": {"label": "1", "order": 1, "volume": "1", "name": "Название главы"},
  "source_note": "Источник и основание для загрузки",
  "credits": "Автор перевода / команда",
  "rights_confirmed": true,
  "preparation_method": "machine_assisted",
  "human_reviewed": false
}
```

3. Из checkout Qmanga запусти:

```sh
python3 scripts/qmanga-upload.py --base-url https://qmanga.maqsatto.dev \
  --token-file "$HOME/.config/qmanga/upload.token" \
  --manifest chapter.json --source ./chapter.cbz --publish
```

Глава публикуется от аккаунта владельца токена без модерации. Без `--publish` остаётся подготовленным черновиком; `--submit` использует обычную модерацию. `human_reviewed: true` указывай только после реальной проверки человеком.

**Продолжить после сбоя:** повтори ту же команду, сохранив `chapter.qmanga-state.json`. Не меняй входные страницы или manifest у начатого импорта. Сортировка естественная (`1,2,10`); 300 страниц, 25 МиБ/файл, 500 МиБ/главу. Чужие страницы черновика клиент не удаляет.

**Напрямую через API:** `Authorization: Bearer <TOKEN>`, без cookie/CSRF. `GET /api/v1/auth/me`; создание тайтла `POST /api/v1/admin/titles`; глав и проектов — `/admin/titles/{id}/chapters` и `/projects`. Загрузка и версии описаны в [контракте](PREPARED-PUBLICATION-IMPORT.md); прямая публикация — `POST /api/v1/studio/revisions/{id}/publish` с `{"version":N}`. Просроченный/отозванный токен или потеря прав админа прекращают доступ.
