# Qmanga Component First Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Собрать компактный mobile-first интерфейс существующего Qmanga из проверенных компонентов, сохранив цвета и сценарии, с Kazakh-first / Russian UI.

**Architecture:** Общие примитивы находятся в shared/presentation, продуктовые компоненты — в presentation своего bounded context. Сценарии и хранилища остаются за существующими application ports; локаль передаётся через отдельный контекст layout и platform adapters. Сначала завершить Tasks 1–3, только затем переносить страницы.

**Tech Stack:** Текущие Bun, Svelte 5, SvelteKit, Tailwind 3, TypeScript и Alpine Docker из lockfile и существующих образов; без обновления зависимостей и новой UI-библиотеки.

**Spec:** [2026-10-17-component-first-redesign-design.md](../specs/2026-10-17-component-first-redesign-design.md), подтверждена владельцем сообщением «лес го» после публикации спецификации.

**Execution:** Native: реализовать самостоятельно в текущем checkout `outputs/qmanga`, без AGY, новых веток и worktree. После реализации — независимое итоговое ревью по Superpowers. Этот план требует проверки владельцем перед началом реализации; не запрашивать повторный выбор исполнителя.

## Global Constraints

- Существующие роли и значения цветов сохраняются: dark `#0d1117 / #161c25 / #202835 / #29323e / #edf2f5 / #96a2b2 / #b4ee69 / #17220d`; light `#f6f8f3 / #ffffff / #edf1e8 / #d9dfd3 / #172217 / #65715f / #486c1d / #ffffff`, порядок ролей — как в спецификации.
- Системный шрифт с казахской кириллицей, без загрузки новых шрифтов. Основной текст и ввод — 16px, метаданные — 13–14px, подписи мобильной навигации — 12px, заголовок страницы — 24px на телефоне и 28px на широком экране.
- Рабочая шкала отступов: 4, 8, 12, 16, 24, 32px; скругления: 6px для обложек, 8px для контролов, 12px для отдельных панелей.
- Шапка высотой 56px; поля страницы 16px на телефоне / 24px на десктопе; основной контейнер максимум 1200px. Каталог: 2 / 3 / 4 / 5 колонок при ширинах 320 / 640 / 768 / 1024px, обложки 7:10.
- Первый визит — `kk`, независимо от языка устройства. Видимые названия: «Қазақша / Русский». Использовать «Аккаунт», «Команда», «Каталог» и обычные казахские действия; «Профиль» — только там, где действительно есть профиль.
- Цели интерактивных элементов минимум 44×44px. Текст 4.5:1, крупный текст и необходимые границы/индикаторы 3:1. Плотность не оправдывает потерю целей или фокуса.
- Сетевой HTML с локалью из cookie получает `Cache-Control: private, no-store`; `/offline-reader` остаётся общей prerender-оболочкой `kk` без аккаунтных данных. Медиа и статические файлы сохраняют текущую cache policy.
- Все обложки и страницы остаются за Go media gateway. Без новых API/полей БД, форматов upload, профилей, лент, поиска глав и других функций из backlog.
- Локаль не меняет reading URL, release/team ID, progress key, качество или выбор перевода. Справа остаётся «вперёд» по умолчанию, включая RTL-источники.
- Ни fetch, ни cookies, ни storage не переносить в примитивы или feature-компоненты. Не добавлять inline code comments. В Git и доказательства не включать пароли, email пользователей, личную историю или исходные страницы манги.

## Review Focus

1. Длинный казахский/русский текст и 200% zoom: сохраняются название, доступное имя, 44px действия и отсутствие page overflow — Tasks 1–3, 10.
2. Одновременные SSR `kk` / `ru` и старый offline cache: нет общего изменяемого locale store, утечки языка/аккаунта или кэширования персонального HTML — Task 4.
3. Неизвестный Go error code, HTML вместо JSON и потеря сети: понятная локализованная ошибка, а не сырой текст и не пустой результат — Tasks 5–9.
4. Данные до редизайна в IndexedDB: главы и прогресс остаются читаемыми; старые `message` не являются источником перевода — Task 8.
5. Смена локали во время чтения/upload: не пересоздаёт контроллер/очередь и не выполняет запрос/действие повторно — Tasks 8–10.

## Общая проверка и коммиты

Все пути ниже относительно корня репозитория. `W` в командах не используется: рабочая директория web-команд внутри контейнера — `/app`, соответствующая `apps/web`.

Каждый Task проходит red → green: сначала целевой тест должен упасть из-за отсутствующего поведения, затем минимальная реализация, целевой тест повторно. Для чисто визуальных свойств использовать явные browser assertions и скриншоты, не выдавать поиск строки в CSS за визуальную проверку. После каждого Task: `docker compose build web` с architecture check, unit/UI tests, Svelte check и production build; затем `git diff --check`, коммит только его файлов и `git push origin main`. При неуспехе не отмечать Task завершённым.

## Task 1 — интерактивные примитивы и изолированный UI preview

**Create:** `apps/web/src/lib/shared/presentation/styles/tokens.css`; `apps/web/src/lib/shared/presentation/ui/types.ts`; `Button.svelte`, `ActionLink.svelte`, `Icon.svelte`, `IconButton.svelte`, `Field.svelte` в том же `ui/`; `apps/web/ui-preview/index.html`, `main.ts`, `Preview.svelte`; `apps/web/vite.ui-preview.config.ts`, `tsconfig.ui-preview.json`; `apps/web/tests/ui/setup.ts`, `markup.ts`, `Controls.fixture.svelte`, `controls.test.ts`.

**Modify:** `apps/web/package.json`, `infra/web.Dockerfile`, `compose.yaml` — только test/preview tooling. Основные экраны и глобальное оформление пока не переносить.

**Interfaces:** `ButtonProps` расширяет native button attributes: `variant: 'primary'|'secondary'|'ghost'|'danger'`, `busy?: boolean`, `children: Snippet`; default `type='button'`. `ActionLink` принимает native anchor attributes и вариант, но не `disabled`. `IconName` — union реально нужных собственных SVG. `IconButton` требует `name: IconName` и `label: string`. `Field` принимает `id`, `label`, `hint?`, `error?`, `required?`, `disabled?`, `children: Snippet<[FieldControl]>`; `FieldControl` содержит `id`, `required`, `disabled`, `'aria-describedby'?`, `'aria-invalid'?`. ID передаётся на native control через snippet, а не генерируется случайно.

- [ ] **Harness:** `tests/ui/setup.ts` регистрирует Bun runtime plugin для `.svelte`: `svelte/compiler.compile` с `generate:'server'`, разрешение `$lib` в `src/lib`. `markup.ts` использует `svelte/server.render` и предоставляет `elements(body:string,selector:string):Promise<Record<string,string>[]>` через Bun `HTMLRewriter`. Запуск `bun test --preload ./tests/ui/setup.ts ./tests/ui`; не подменять SSR ручной генерацией HTML. API: [Bun plugins](https://bun.sh/docs/runtime/plugins), [Svelte compiler](https://svelte.dev/docs/svelte/svelte-compiler), [Svelte server render](https://svelte.dev/docs/svelte/svelte-server).
- [ ] **Red:** в `controls.test.ts` отрендерить настоящую Svelte fixture и проверить: default button type, explicit submit, busy ⇒ disabled + `aria-busy`, сохранение accessible label, ссылка без вложенного button, icon `aria-hidden`, связь label/control/hint/error, экранирование `<script>` в тексте. После создания harness тест должен упасть на импорте ещё отсутствующего `Button.svelte`, а не на отсутствии loader.

```ts
test('controls preserve native semantics', async () => {
 const html = render(ControlsFixture).body;
 expect((await elements(html, '[data-case="default"]'))[0].type).toBe('button');
 expect((await elements(html, '[data-case="busy"]'))[0]).toMatchObject({'aria-busy':'true', disabled:''});
 expect(await elements(html, 'a button')).toHaveLength(0);
 expect(await elements(html, 'script')).toHaveLength(0);
});
```
- [ ] **Implement:** добавить указанные компоненты и токены из спецификации; danger различается текстом/контуром без новой палитры. Поля и кнопки — минимум 44px, фокус — существующий акцент. Никакого курсора `wait` на disabled; busy не меняет ширину контрола.
- [ ] **Preview:** отдельный Vite entry `ui-preview`, `svelte({configFile:false})` с текущим preprocess, `$lib` alias, порт 3001. Только общий CSS/tokens и компоненты; без импорта SvelteKit routes. Надпись «UI preview», явные демонстрационные подписи KK/RU, переключение обеих тем, кнопка-счётчик для проверки одного callback. Preview не попадает в production build entry.
- [ ] **Tooling:** `test:ui` запускает harness; `ui:preview` — отдельный Vite config. `check:ui-preview` использует `svelte-check --tsconfig ./tsconfig.ui-preview.json`; этот config расширяет текущий и включает preview, Svelte fixtures и lib, не только Kit routes. В `infra/web.Dockerfile` добавить `bun run test:ui` и после существующего `bun run check` — `bun run check:ui-preview`; существующий unit command расширить до `bun test src/lib/features src/lib/shared src/lib/server`, чтобы Task 4 server tests тоже входили в gate. В Compose добавить `web-ui` с профилем `tools`, build target `build`, bind `./apps/web:/app`, anonymous `/app/node_modules`, портом только `127.0.0.1:3001:3001`, командой `bun run ui:preview --host 0.0.0.0`. Production runtime по-прежнему копирует только build/package.json.
- [ ] **Green:** `docker compose --profile tools run --rm web-ui bun run test:ui`; открыть preview, проверить фокус/Enter/submit, disabled/busy без повторного callback, 320/360px и две темы. Production `/ui-preview` не должен показывать gallery. Полный build gate, коммит `feat(ui): add accessible controls and isolated component preview`.

## Task 2 — состояния, заголовки и навигационные примитивы

**Create:** `apps/web/src/lib/shared/presentation/ui/Badge.svelte`, `SectionHeading.svelte`, `StateMessage.svelte`, `SettingsRow.svelte`, `Pager.svelte`; `apps/web/tests/ui/Content.fixture.svelte`, `content.test.ts`.

**Modify:** `apps/web/src/lib/shared/presentation/ui/types.ts`, `apps/web/ui-preview/Preview.svelte`.

**Interfaces:** `Badge {label, tone?:'neutral'|'accent'|'warning'|'danger'}` — span, не action. `SectionHeading {level:1|2|3,title,description?,actions?:Snippet}`. `StateMessage {kind:'empty'|'loading'|'error'|'offline'|'success',title,description?,actions?:Snippet}`. `SettingsRow {controlId,label,description?,children:Snippet}`. `PageControl = {label:string;href:string}|{label:string;action:()=>void;busy?:boolean;disabled?:boolean}`; `Pager {previous?:PageControl,next?:PageControl,label:string}`. Нет автоматических запросов или придуманных переходов.

- [ ] **Red:** fixture проверяет h1/h2/h3, некликабельный Badge, `alert` только для ошибки, status для loading/offline/success, отсутствие live-region у empty, отсутствие пустого контейнера actions, связь SettingsRow с native select. Pager сохраняет href ссылки, использует button для callback, блокирует busy/disabled. Запуск целевого `content.test.ts` должен упасть на новых импортах.

```ts
test('content states do not announce an empty result as a failure', async () => {
 const html = render(ContentFixture).body;
 expect(await elements(html, '[data-case="empty"] [role="alert"]')).toHaveLength(0);
 expect(await elements(html, '[data-case="error"] [role="alert"]')).toHaveLength(1);
 expect((await elements(html, '[data-case="next-link"]'))[0].href).toBe('/catalog?cursor=test');
});
```
- [ ] **Implement:** собрать на примитивах Task 1; стили различают строки, метки и состояния, не делают из каждого блока panel. У SettingsRow контроль переносится при нехватке ширины; перенос иконок не обрезает подпись.
- [ ] **Green:** целевой UI test, preview с длинными текстами на 320/360/1280px, минимум 44px у действий; общий build gate. Коммит `feat(ui): add content states and settings primitives`.

## Task 3 — продуктовые компоненты без переноса страниц

**Create:** `apps/web/src/lib/features/catalog/presentation/CoverCard.svelte`, `ChapterRow.svelte`; `apps/web/src/lib/features/reading/presentation/LibraryRow.svelte`; `apps/web/src/lib/features/identity/presentation/AuthLayout.svelte`; `apps/web/src/lib/shared/presentation/layout/AppHeader.svelte`, `MobileNavigation.svelte`, `navigation.ts`; `apps/web/tests/ui/Product.fixture.svelte`, `product.test.ts`; `apps/web/src/lib/shared/presentation/layout/navigation.test.ts`.

**Modify:** `apps/web/ui-preview/Preview.svelte` — добавить вымышленные, явно помеченные UI samples без artwork/аккаунтов и сетевых запросов. `Product.fixture.svelte` принимает `section?:'library'|'catalogue'|'all'` и `releaseId?:string`, чтобы проверять LibraryRow без посторонних chapter links.

**Interfaces:** `CoverCard {title:Summary,labels:{kind,chapters,demo,noCover},priority?:boolean}`; `ChapterRow {chapter:Chapter,labels:{read,volume,chapter},dateLabel:(iso:string)=>string}`; `LibraryRow {entry:LibraryEntry,labels:{statuses:Record<string,string>,unseen:string,favorite:string,continue:string,remove:string},onupdate:(entry:LibraryEntry)=>void,onremove:()=>void,busy?:boolean}`. Все labels — строки/типизированные словари, не locale store. `AuthLayout {title,description?,children:Snippet,footer?:Snippet}`. `NavigationItem {path,label,icon:IconName}`; `activeDestination(pathname:string):string|null`; обе навигации принимают items + activePath + доступные имена извне. AppHeader отдельно принимает accountHref/accountLabel и homeLabel, не модель аккаунта.

- [ ] **Red:** product tests: cover 7:10 с width/height, полный accessible title при обрезанном отображении, без cover URL нет broken img, demo явно виден, lazy вне приоритета; ссылки команды/чтения отдельные; Continue в LibraryRow только при releaseId; email не нужен ни одному shell-компоненту. Navigation tests: `/`, `/catalog`, `/titles/x`, `/library`, `/notifications`, `/settings`, `/read/x` → правильный раздел или null для reader.

```ts
test('navigation keeps the reader outside the site shell', () => {
 expect(activeDestination('/read/test')).toBeNull();
 expect(activeDestination('/titles/test')).toBe('/catalog');
 expect(activeDestination('/notifications')).toBe('/notifications');
});
```
- [ ] **Implement:** callbacks без HTTP/storage и без правил выбора перевода. У LibraryRow обновление статуса/избранного передаётся владельцу; никаких новых read models или запросов за обложками. AppHeader 56px; MobileNavigation пять элементов с `aria-current`, 44px, safe-area padding. Отдельный desktop вариант оставляет все пять направлений доступными.
- [ ] **Green:** целевые тесты, preview на 320/360/768/1280px, два языка/темы, нет page overflow и вложенных ссылок. Общий build gate; коммит `feat(ui): add library catalogue and navigation components`.

**Component gate:** Tasks 1–3 зелёные, preview осмотрен, заметные дефекты устранены. Лишь после этого начать перенос рабочих экранов.

## Task 4 — изолированная локаль и офлайн контракт

**Create:** `apps/web/src/lib/shared/domain/locale.ts`, `locale.test.ts`; `apps/web/src/lib/shared/presentation/i18n/context.ts`, `format.ts`, `format.test.ts`, `common.ts`; `apps/web/src/lib/features/preferences/application/locale.ts`, `locale.test.ts`; `apps/web/src/lib/features/preferences/infrastructure/browser-locale-store.ts`, `browser-locale-store.test.ts`; `apps/web/src/lib/server/locale.ts`, `locale.test.ts`; `apps/web/src/hooks.server.ts`, `apps/web/src/app.d.ts`; `apps/web/tests/ui/Locale.fixture.svelte`, `locale.test.ts`.

**Modify:** `apps/web/src/app.html`, `apps/web/src/routes/+layout.server.ts`, `+layout.svelte`; `apps/web/src/lib/composition/browser.ts`; `apps/web/src/routes/read/[id]/+page.server.ts`; `apps/web/src/service-worker.ts` — только если требуется закрепить общий offline fallback, без изменения медиа policy.

**Interfaces:** `Locale='kk'|'ru'`, `localeFrom(value:unknown):Locale`. `LocaleStore {read():string|null;save(value:Locale):void;apply(value:Locale):void}` — чистый port preferences. `provideLocale(initial:Locale,onChange:(locale:Locale)=>void):LocaleContext`, `useLocale():LocaleContext`; `LocaleContext {locale:Readable<Locale>;choose(value:Locale):void}`. `common(locale:Locale):CommonMessages` возвращает один из immutable typed dictionaries, не читая context автоматически. `formatNumber(locale:Locale,value:number):string`, `formatDate(locale:Locale,value:string|number):string` с локализованным fallback для invalid date. Cookie `qmanga_locale`, storage key `qmanga.locale`, Path `/`, SameSite Lax, max-age 31536000, Secure при HTTPS. Cookie не содержит аккаунтных данных.

- [ ] **Red:** default/invalid/uppercase/undefined ⇒ kk, явный ru ⇒ ru; ignore navigator language. Port doubles: save/apply при смене, storage exceptions не прерывают выбор. Два параллельно SSR-rendered Locale fixtures не влияют друг на друга; числа/даты KK/RU, invalid date не становится Epoch. Целевые тесты должны упасть на отсутствующих модулях.

```ts
test('locale is Kazakh unless explicitly supported', () => {
 expect(localeFrom(undefined)).toBe('kk');
 expect(localeFrom('ru-RU')).toBe('kk');
 expect(localeFrom('RU')).toBe('kk');
 expect(localeFrom('ru')).toBe('ru');
});
```
- [ ] **Implement:** request-scoped `locals.locale` из cookie, отдельный Svelte context на каждый layout; hooks меняет только `<html lang>` через фиксированный marker app.html. Офлайн route всегда SSR/prerender kk и user=null, без чтения request cookie при prerender. Обычная сеть использует SSR cookie; локальное восстановление применяется только offline-оболочке. `choose` меняет store и вызывает injected adapter, без goto/invalidate/remount страницы.
- [ ] **Cache:** hooks централизованно ставит `private, no-store` на локализованный HTML и соответствующие SvelteKit data responses; убрать конфликтующие public reader headers, не трогать `/api/v1/media` и статику. Проверить сетевые и client-navigation ответы. Общий `/offline-reader` не локализуется по cookie и не содержит user HTML. Аккаунтная очистка при logout не стирает общий выбор языка.
- [ ] **Green:** targeted unit/UI tests; HTTP проверки с kk/ru cookies одновременно и без cookie, проверка `<html lang>`, headers и offline prerender. Browser: storage forbidden, выбор в текущей вкладке продолжает работать. Общий build gate; коммит `feat(ui): add request isolated Kazakh first localization`.

## Task 5 — оболочка, аккаунт и формы

**Create:** `apps/web/src/lib/shared/domain/failure.ts`, `failure.test.ts`; `apps/web/src/lib/shared/presentation/i18n/failure.ts`, `failure.test.ts`; `apps/web/src/lib/features/identity/presentation/messages.ts`; `apps/web/src/lib/features/preferences/presentation/messages.ts`; `apps/web/src/lib/features/identity/application/auth.test.ts`; `apps/web/src/lib/features/identity/infrastructure/http-auth-gateway.test.ts`; `apps/web/src/routes/+error.svelte`; `apps/web/tests/ui/auth.test.ts`, `shell.test.ts`.

**Modify:** `apps/web/src/app.css`, `apps/web/src/routes/+layout.svelte`; `apps/web/src/lib/features/preferences/presentation/Settings.svelte`; `apps/web/src/lib/features/identity/application/auth.ts`, `infrastructure/http-auth-gateway.ts`, `presentation/AuthForm.svelte`; `apps/web/src/lib/shared/infrastructure/api-client.ts`, `api-client.test.ts`; `apps/web/src/routes/login/+page.svelte`, `register/+page.svelte`, `forgot-password/+page.svelte`, `reset-password/+page.svelte`, `verify-email/+page.svelte`, `settings/+page.svelte`.

**Interfaces:** `AppFailure extends Error {readonly code:string;readonly status:number}`; constructor `(code:string,status=0)`; `.message` — технический code, никогда выводимый как UI. `failureCode(value:unknown):string` возвращает code только у AppFailure, иначе `unknown`; presentation resolver использует whitelist словаря и safe fallback. ApiFailure остаётся совместимым subtype с status/payload для текущих инфраструктурных потребителей. `AuthResult {ok:boolean;code:AuthNoticeCode;account?:Account}`; `AuthNoticeCode` — union `none|invalid_email|invalid_password|login_ok|register_sent|reset_sent|verification_sent|verified|password_changed|logged_out|validation|credentials|expired|conflict|unavailable|unauthorized|forbidden|csrf|origin|invalid_json|rate_limit|network|unknown`. Неизвестный серверный код нормализуется в unknown, success code определяется действием, не payload.message. AuthForm дополнительно принимает `token?:string`, `onAuthenticated?:()=>Promise<void>`; Settings — `onLoggedOut?:()=>Promise<void>`. Навигация/token из `$app` остаются в routes и передаются явно, чтобы Svelte SSR fixtures не зависели от виртуальных модулей SvelteKit. Существующие auth методы/правила/CSRF остаются теми же.

- [ ] **Red:** malformed JSON/network/timeout/unknown server code → safe code без раскрытия HTML/English/raw message; известные auth codes и success messages локализуются KK/RU; credential limits остаются 10–128. Shell fixture показывает все направления и аккаунт/вход без email, reader bypass без двойной навигации. Auth fixture: прямые заголовки, label/autocomplete/required, password reset/verify links и explicit submit. Целевые тесты падают на старых строковых результатах/структуре.

```ts
test('API errors do not expose arbitrary server text', async () => {
 const api = new ApiClient(async () => Response.json({error:{code:'unrecognized',message:'secret'}},{status:422}));
 await expect(api.call('/api/v1/titles')).rejects.toMatchObject({code:'unknown',status:422});
 expect(failureCode(new Error('secret'))).toBe('unknown');
});
```
- [ ] **Implement:** context-specific `messages(locale)` возвращает типизированные словари; общие слова — common, бизнес-статусы — свой контекст. Заменить shell на Task 3, формы — Task 1 + AuthLayout. Подключить tokens, применить типографику и mobile/desktop container. Удалять старые global selectors только когда больше нет потребителя. У SettingsRow темы/языка реальные текущие handlers; выбор языка «Қазақша / Русский».
- [ ] **Error page:** 404/503 и остальные статусы через context/SSR locale, без raw message; фиксированные доступные переходы домой/каталог. Переход к экрану ошибки не должен терять язык. Auth login/logout/verify/reset сохраняют текущие gateway calls и invalidate/goto только там, где уже были.
- [ ] **Green:** целевые тесты; оба языка/темы на auth/settings в браузере, email и пароль не включать в screenshots. Проверить mobile keyboard overlap и 200% zoom. Общий build gate; коммит `refactor(ui): compose account forms and site navigation`.

## Task 6 — каталог, главная и страницы тайтла

**Create:** `apps/web/src/lib/features/catalog/presentation/messages.ts`, `Home.svelte`; `apps/web/tests/ui/catalogue.test.ts`, `title.test.ts`.

**Modify:** `apps/web/src/lib/features/catalog/presentation/Catalogue.svelte`, `Title.svelte`, `PublicTeam.svelte`, `AdminCatalogue.svelte` только импорты перенесённых status labels; `apps/web/src/lib/features/catalog/domain/catalog.ts`, `application/catalog.ts`, `catalog.test.ts`; `apps/web/src/routes/+page.server.ts`, `+page.svelte`, `catalog/+page.server.ts`, `titles/[slug]/+page.server.ts`, `titles/[slug]/+page.svelte`, `teams/[slug]/+page.server.ts`.

**Interfaces:** enums/status IDs не меняются; `kindNames`/status display names переходят из domain в catalog presentation dictionary. `loadCatalogue` возвращает `code:''|'catalogue_unavailable'` вместо готового `message`, сохраняя available/page/genres. Home использует тот же `loadCatalogue(createPublicCatalogue(fetch),{sort:'updated'})`, без нового порта/API. `Title` получает optional `library?:Snippet` и выводит его после header до описания/глав; route передаёт существующий LibrarySelection, не создаёт второй instance.

- [ ] **Red:** home fixtures: реальные items видны, empty не равен failed, updatedAt не назван датой новой главы; каталог сохраняет GET query names/filter values/cursor links; cover grid использует Task 3. Title fixture проверяет library snippet перед chapters, в браузере — один LibrarySelection и существующий Discussion; нет новой неработающей CTA. Дополнить существующий application test передачей sort/empty/failure, а выбор updated в настоящей home composition проверить на живой странице, не фиктивным вызовом внутри теста.

```ts
test('catalogue distinguishes an empty result from an unavailable service', async () => {
 const gateway = {list:async () => ({items:[],nextCursor:null}),genres:async () => []} as unknown as CatalogGateway;
 expect((await loadCatalogue(gateway, {})).available).toBe(true);
 gateway.list = async () => {throw new Error('secret');};
 expect(await loadCatalogue(gateway, {})).toMatchObject({available:false,code:'catalogue_unavailable'});
});
```
- [ ] **Implement:** убрать hero/orbit/health decoration; компактный поиск GET `/catalog` и настоящие CoverCard. Сохранить demo labels/URL, pagination и существующие filter/details. Перевести текущие route error codes в безопасные localized errors, обложки не обходят Go. На тайтле использовать ChapterRow, компактную metadata hierarchy и существующие команды/описание/aliases.
- [ ] **Green:** unit/UI tests, проверка реального непустого каталога и отключённого API, 320/360/768/1280px с grid 2/2/4/5; дополнительно 640px ⇒ 3. Общий build gate; коммит `refactor(ui): compose catalogue home and title screens`.

## Task 7 — библиотека, обсуждения и уведомления

**Create:** `apps/web/src/lib/features/reading/presentation/messages.ts`, `community/presentation/messages.ts`, `notifications/presentation/messages.ts` под тем же `apps/web/src/lib/features/`; `apps/web/tests/ui/library.test.ts`, `community.test.ts`, `notifications.test.ts`.

**Modify:** `apps/web/src/lib/features/reading/presentation/Library.svelte`, `LibrarySelection.svelte`; `apps/web/src/lib/features/community/presentation/Discussion.svelte`, `ReportForm.svelte`, `Reports.svelte`, `domain/content.ts` только перенос category display names; `apps/web/src/lib/features/notifications/presentation/Notifications.svelte`; `apps/web/src/lib/features/community/infrastructure/http-community-gateway.ts`, `notifications/infrastructure/http-notifications-gateway.ts`, `reading/infrastructure/http-reading-gateway.ts`, `identity/infrastructure/http-account-scope.ts` — ошибки/notice mapping, не изменение API calls.

**Interfaces:** существующие gateways, LibraryEntry/Selection/Discussion/Notification неизменны; callbacks LibraryRow из Task 3 вызывают текущие update/remove. Mapper failures становятся AppFailure по Task 5; notices об успешном действии хранят код, не snapshot локализованной строки. Даты/counts — formatter Task 4. Пользовательские body/reason/title/team name не переводятся. Notification heading берётся по kind `new_release|comment_reply`, unknown kind получает общий localized heading; сохранённый legacy body показывается как исходное содержимое с `lang="kk"`, не парсится и не переписывается. Текущее API не отдаёт отдельно имя/номер главы: не обещать перевод исторического body и не добавлять N+1 metadata fetch ради UI.

- [ ] **Red:** LibraryRow сохраняет status/favorite/remove/continue и приглашение в аккаунт, no auto guest import; конфликт показывает обе существующие альтернативы. Discussion: spoiler/reply/rating/report не пропадают, email не становится author name. Notifications: read/unread и реальные ссылки сохраняются. Unknown/malformed gateway data получают localizable fallback, не raw `.message`; переключение RU не меняет пользовательский body.

```ts
test('library rows do not invent a continuation URL', async () => {
 const html = render(ProductFixture, {props:{section:'library',releaseId:''}}).body;
 expect(await elements(html, 'a[href^="/read/"]')).toHaveLength(0);
});
```
- [ ] **Implement:** собрать lists/forms из компонентов, сохранить owner controllers и проверку account change. LibrarySelection выше длинного chapter list; пустые/ошибочные/гостевые/loading состояния отдельные. Не добавлять профили/вкладки/новый порядок глав.
- [ ] **Green:** UI tests и существующие guest-import/progress tests; браузер reader-account: library save/edit/remove, guest conflict, comment/spoiler/report и mark read. Общий build gate; коммит `refactor(ui): compose library community and notification screens`.

## Task 8 — читалка и старые офлайн данные

**Create:** `apps/web/src/lib/features/offline/presentation/messages.ts`; `apps/web/src/lib/features/offline/infrastructure/download-record.ts`, `download-record.test.ts`; `apps/web/tests/ui/reader.test.ts`, `offline.test.ts`.

**Modify:** `apps/web/src/lib/features/reading/presentation/Reader.svelte`; `reading/application/progress.ts`, `reader.ts`, `scheduler.ts`, их `progress.test.ts`, `scheduler.test.ts`; `reading/infrastructure/browser-image-loader.ts`, `http-reader-gateway.ts`; `apps/web/src/lib/features/offline/application/downloads.ts`, `downloads.test.ts`, `infrastructure/browser-download-storage.ts`, `saved-image-loader.ts`, `saved-image-loader.test.ts`, `presentation/Saved.svelte`, `OfflineReader.svelte`; `apps/web/src/lib/features/reading/presentation/messages.ts`; `apps/web/src/routes/read/[id]/+page.server.ts` только localized error mapping; `apps/web/src/routes/offline-reader/+page.svelte` передаёт `releaseId?:string` вместо чтения `$app/stores` внутри OfflineReader.

**Interfaces:** у ReadingSession заменить готовый `message` на `notice:ReadingNoticeCode`; union — по существующим состояниям, включая accountChanged/conflict/localSaved/cloudSaved/networkFailure. MediaFailure сохраняет terminal/retryAfter; добавить `code` с default `media_unavailable`, не менять retry timing и budget. ImageState и Download получают `notice` вместо отображения сырого текста; `downloadFrom(value:unknown):Download` допускает legacy `message`, игнорирует его как UI-источник и выводит код по status (`none` для ready, `download_partial` для partial, `download_failed` для error). Release/revision/tier/files/checksum/position остаются без миграции ключей и без очистки storage. Не распознавать причину старой ошибки по казахскому тексту.

- [ ] **Red:** текущие navigation/layout/scheduler/progress тесты остаются; дополнить terminal/retryAfter сохранением при новой error mapping. Старый Download с `message`, но без notice читается/проверяется/открывается, а bytes/files/progress не меняются. И kk, и ru notices реагируют на смену языка после уже случившейся ошибки. Offline failure об удалённом/неполном файле различим от сетевой ошибки.

```ts
test('legacy ready downloads survive localization', () => {
 const restored = downloadFrom({...legacyDownload,status:'ready',message:'Ескі хабар'});
 expect(restored).toMatchObject({id:legacyDownload.id,files:legacyDownload.files,bytes:legacyDownload.bytes,notice:'none'});
 expect(restored.manifest.releaseId).toBe(legacyDownload.manifest.releaseId);
});
```
- [ ] **Implement:** только visual primitives и labels/panels; текущие callbacks/обработчики остаются. Язык в настройках reader использует Task 4, без remount ImageQueue/ReadingSession. Ошибки adapters — codes; scheduler не анализирует локализованные строки. Читалка сохраняет back/title/chapter, page counter, mobile bottom controls, desktop edges/drawer, focus/escape; original quality остаётся существующим действием.
- [ ] **Green:** targeted unit/UI tests, мышь/клавиши/мобильные taps справа next и слева previous независимо RTL, границы страниц/глав, перезагрузка и сохранение позиции. Сохранить главу, переключить RU, отключить сеть и открыть saved reader: тот же release/позиция, локализованные панели, без account HTML в cache. Скриншоты обеих тем; малые высоты/landscape и 200% zoom. Общий build gate; коммит `refactor(ui): align reader and offline presentation without policy changes`.

## Task 9 — рабочие кабинеты

**Create:** `apps/web/src/lib/features/teams/presentation/messages.ts`, `publishing/presentation/messages.ts`; `apps/web/src/lib/features/teams/application/teams.test.ts`; `apps/web/src/lib/features/publishing/application/publishing.test.ts`; `apps/web/tests/ui/workspaces.test.ts`.

**Modify:** `apps/web/src/lib/features/teams/presentation/Studio.svelte`, `TeamApprovals.svelte`, `application/teams.ts`, `infrastructure/http-teams-gateway.ts`; `apps/web/src/lib/features/publishing/presentation/Projects.svelte`, `PublishingStudio.svelte`, `Moderation.svelte`, `application/publishing.ts`, `infrastructure/http-publishing-gateway.ts`; `apps/web/src/lib/features/catalog/presentation/AdminCatalogue.svelte`, `catalog/presentation/messages.ts`, `catalog/application/management.ts`; `apps/web/src/routes/teams/invitation/+page.svelte`.

**Interfaces:** текущие TeamsGateway/PublishingGateway/CatalogManagementGateway и разрешения неизменны. Валидация/отмена UploadQueue выдаёт AppFailure с кодом, без изменения limits/natural order/retry/upload key/position. Source processing failure без безопасного стабильного кода показывается локализованным общим пояснением; note/reason переводчика/модератора остаётся пользовательским текстом.

- [ ] **Red:** tests фиксируют существующие team validation limits, owner/editor/moderator choices, natural upload order и stop/retry без дубликатов; fixtures сохраняют approve/invite/submit/publish/reject/withdraw и catalog editor. Неизвестный processing failure не выводит raw exception. Переключение локали в процессе upload не вызывает второй run/upload.

```ts
test('UI locale cannot alter the publishing natural order', () => {
 expect(naturalOrder([{name:'10.webp'},{name:'2.webp'},{name:'1.webp'}]).map(item => item.name)).toEqual(['1.webp','2.webp','10.webp']);
});
```
- [ ] **Implement:** SectionHeading + рабочие rows + Field + Button/StateMessage, существующие state machines не переносить в компоненты. На телефоне controls переносятся и сохраняют 44px. Сохранить роли/noindex и все существующие действия, без профилей, аватаров, ZIP/CBZ и новых API.
- [ ] **Green:** целевые тесты; локальные owner/editor/admin аккаунты: приглашение/заявка, существующий upload/reorder/review и отказ reader в служебных действиях. Файлы тестовой манги остаются вне Git. Общий build gate; коммит `refactor(ui): compose translator and moderation workspaces`.

## Task 10 — итоговая проверка редизайна

**Create:** `apps/web/tests/ui/localization.test.ts`; `apps/web/scripts/check-ui-contrast.mjs`; `docs/ui/REDESIGN-VALIDATION.md`.

**Modify:** `apps/web/package.json`, `infra/web.Dockerfile` — включить `check:ui-contrast` в gate; `docs/ui/DESIGN-REVIEW.md`, `docs/VALIDATION.md`, `docs/PROGRESS.md` — только реальные результаты/остаточные ограничения. Удалить из `apps/web/src/app.css` уже неиспользуемые hero/orbit/eyebrow/auth legacy styles, не общую поддержку текущих controls.

**Interfaces:** `check:ui-contrast` проверяет фактические пары token colors для text/muted/accent-ink/focus/control outline, пороги 4.5/3.0. Если роль не проходит, поменять её сочетание внутри сохранённой палитры, не «ослабить» threshold. `localization.test.ts` сравнивает ключи и типы kk/ru словарей и динамические notices каждого контекста; пользовательский контент не ищется как «непереведённый UI».

- [ ] **Red:** исключённый RU key, контрастная пара ниже порога и русская locale fixture со старым hardcoded notice должны давать failing tests. Затем закрыть пробелы и восстановить валидные значения; не оставлять намеренно сломанные fixture variants в production.

```ts
test('common dictionary keys cover both supported locales', () => {
 expect(Object.keys(common('kk')).sort()).toEqual(Object.keys(common('ru')).sort());
});
```
- [ ] **Browser matrix:** home/catalog/title/library/auth/settings/notifications/saved/studio/moderation/admin/online-reader/offline-reader; размеры 320, 360, 768, 1280px; KK/RU и dark/light. Отдельно long content, populated/empty/error/offline, keyboard focus, 200% zoom, reduced motion, virtual keyboard и safe area. Измерить реальные target rects и document overflow; не считать screenshot доказательством callback/HTTP поведения.
- [ ] **Network/regression:** cold reload, client navigation и параллельные cookies для SSR; все медиа через Go; offline cache без аккаунтного HTML; смена языка без сброса release/progress/команды и без повторного upload. Полный `docker compose build web`, `docker compose up -d --wait web`, `docker compose --profile tools run --rm go-tests` для финальной межслойной регрессии, health всех сервисов.
- [ ] **Green:** все указанные команды завершаются exit 0, architecture/contrast/unit/UI/Svelte/build без ошибок, сервисы healthy; browser assertions фиксируют ожидаемые действия и размеры. Любое непроверенное условие записать как ограничение, не как pass.
- [ ] **Evidence:** записать свежие результаты/непроверенные случаи в REDESIGN-VALIDATION, screenshots — в соседнем user-facing `outputs/ui-review/`, без приватных аккаунтов/чужой истории. Физический слабый Android не отмечать проверенным по browser emulation; оставить это условием публичной беты.
- [ ] **Review/commit:** независимое итоговое ревью по `superpowers:requesting-code-review`, устранение относящихся к редизайну замечаний, повтор нужных проверок; `git diff --check`, коммит `test(ui): verify bilingual mobile first redesign`, push и чистое дерево. Существующий функциональный backlog не отмечать выполненным.

## Plan Review

- [x] Спецификация покрыта: примитивы 1–2, продуктовые компоненты 3, locale/cache 4, все экраны 5–9, доступность/регрессии 10.
- [x] Порядок владельца сохранён: рабочие страницы не переносятся до component gate.
- [x] Типы и интерфейсы между задачами согласованы; реализация хранится в текущем checkout и не делегируется AGY.
- [x] Пять Review Focus условий закреплены за конкретными тестами/проверками.
- [ ] Владелец проверил этот письменный план перед выполнением.
