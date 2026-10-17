# Qmanga UI direction and review

Updated: 17 October 2026. Status: owner has now requested the general design pass and Kazakh-first/Russian UI. The audit and requirements below are recorded; no general redesign or localization is implemented by this research checkpoint. The proposed composition still needs design confirmation through Superpowers.

## Owner decisions

- Keep green. The owner explicitly clarified that the problem is the composition, not the color. A blue replacement was only a proposal and is not approved.
- Complete functionality first, redesign at the end. That deferral is now lifted by the owner's explicit request to start design; retain working functionality and architecture throughout the visual pass.
- Preserve mobile-first behavior, Kazakh language and familiar reading-site navigation. Senkuro is a workflow/density reference, not permission to copy its assets, branding or private account data.
- Use the reusable design-product-ui skill when the final presentation pass begins.
- Reader control correction: default right-next/left-previous for this audience, including RTL manga. Title reading direction must not mirror mobile or desktop controls; reverse page controls require an explicit manual preference.
- Keep the existing palette, not just green in isolation. The requested change is composition, density and wording, not a replacement theme.
- Kazakh first does not mean terminological purism: use recognizable modern vocabulary rather than formal/literal constructions such as Есептікжазба. Аккаунт and Профиль are appropriate where they describe the actual task. Keep ordinary Kazakh actions, not a random mixture of Russian sentences and Kazakh labels.
- Add complete Russian UI as an explicit alternative. Default to Қазақша before a choice, remember the user's explicit choice, and do not infer Russian from device settings. Switching UI must not translate manga pages, names or comments, change releases, or reset reading progress.

## Observed baseline

Inspected the running home and registration pages at360px, and the existing CSS/components. The home is a promotional hero with a large slogan and decorative desktop Q/orbit. Registration uses a promotional heading, decorative arrows and duplicated brand links. Uppercase tracked labels, repeated rounded containers and excessive empty space compete with task information. Green itself is not a defect.

Fresh audit after functional implementation: at360px the home hero is340px tall and the home still announces that the first stories are coming soon. The catalogue now contains two local test titles with published chapters. The home is not consuming that real content; this is a content-priority/data-state mismatch, not a reason to invent more demonstration content. Catalogue cards are two columns with loaded covers; login still uses a promotional heading and a large vertically spaced form. No credentials or account details are recorded from the browser audit. The temporary viewport override was reset.

## Proposed first design direction

Functional comparison and concrete entry-point gaps: [COMPETITIVE-REVIEW.md](COMPETITIVE-REVIEW.md). In particular, the existing title page lacks the promised library-add action; a visual pass must not hide that missing flow behind new styling.

Recommendation for confirmation: a compact reading library, not a promotional landing page or a social/news dashboard. Keep the current dark/light palette and green selection/action accent. Use restrained headings, real cover grids, compact chapter rows and a clear search/filter hierarchy instead of a hero, decorative Q/orbit, repeated uppercase eyebrows, ornamental arrows and identical large cards for unrelated tasks. Retain five mobile destinations and44px touch targets. At360px, put search and useful reading content in the first screen. Show continuation only when real progress exists and new chapters only from published data; give new visitors the catalogue rather than an empty personal dashboard.

Proposed copy examples: Аккаунт, Профиль, Команда, Каталог, Кіру, Тіркелу, Оқу, Жалғастыру; Russian equivalents are separate translations, not an overlay replacing arbitrary DOM strings. Use the visible language names Қазақша / Русский. The localization scope includes menus, forms, reader, library/offline, studio/admin, accessibility names, validation/errors, dates and counts. It must preserve SSR request isolation, independent domain policies, stable reading URLs and browser/offline behavior; the implementation architecture will be specified after the design direction is confirmed.

## Final-pass brief

Reader amendment: the owner explicitly requests Senkuro-like reading controls now, while the rest of the visual redesign remains deferred. Rechecked its live desktop and360px reader: approximately48px top bar with back/title/chapter, image-first stage, mobile bottom previous/contents/settings/next controls, separate page counter, desktop side controls, full-width mobile settings and chapter drawers, desktop right drawer. Settings include reading mode, theme, fit, numbering and brightness. Qmanga retains green, Kazakh labels and its specified eco/high/original behavior; it does not copy artwork, branding, advertisements or unrelated premium controls. Opening the reference chapter may affect normal reading history; no account details are recorded.

Keep green as an action/selection role; determine exact tokens with both-theme contrast checks. Rework hierarchy and content placement instead of merely recoloring. Home prioritizes actual reading progress and chapter updates. Catalogue prioritizes search, readable cover titles and compact filters. Authentication uses direct task headings such as Тіркелу. Content density must not shrink touch targets or make long Kazakh text unreadable.

Keep five primary mobile destinations. Review at320/360px and desktop, with populated, empty, failed-network and long-content cases. Check44px product touch targets, keyboard focus/overlap, zoom, dark/light contrast and reduced motion. Record screenshots after the real interface is changed. Browser emulation does not establish low-end Android performance.

## Skill evaluation

The pre-skill read-only probe already proposed useful task-first layouts and real data; it is not evidence that every generic design failure was reproduced. Its proposal did not include measurable viewport/contrast/zoom checks. The follow-up using the skill supplied concrete checks and distinguished a project44px target from WCAG baseline. A counterexample correctly retained an explicitly approved cream/serif/Inter design rather than imposing a universal font/color ban. No probe output is an approved Qmanga design or proof of rendered quality.

Sources and limitations: [.agents/skills/design-product-ui/references/sources.md](../../.agents/skills/design-product-ui/references/sources.md).
