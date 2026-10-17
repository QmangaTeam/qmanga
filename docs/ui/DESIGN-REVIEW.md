# Qmanga UI direction and review

Updated: 17 October 2026. Status: redesign deferred until functional MVP, by explicit owner instruction.

## Owner decisions

- Keep green. The owner explicitly clarified that the problem is the composition, not the color. A blue replacement was only a proposal and is not approved.
- Complete functionality first, redesign at the end. Do not interrupt catalogue, publishing or reader implementation with a cosmetic rewrite.
- Preserve mobile-first behavior, Kazakh language and familiar reading-site navigation. Senkuro is a workflow/density reference, not permission to copy its assets, branding or private account data.
- Use the reusable design-product-ui skill when the final presentation pass begins.

## Observed baseline

Inspected the running home and registration pages at360px, and the existing CSS/components. The home is a promotional hero with a large slogan and decorative desktop Q/orbit. Registration uses a promotional heading, decorative arrows and duplicated brand links. Uppercase tracked labels, repeated rounded containers and excessive empty space compete with task information. Green itself is not a defect.

The current empty catalogue is real: content/publishing has not been implemented. Do not fill it with invented ratings, reader counts, copyrighted manga or false reading progress to improve screenshots.

## Final-pass brief

Keep green as an action/selection role; determine exact tokens with both-theme contrast checks. Rework hierarchy and content placement instead of merely recoloring. Home prioritizes actual reading progress and chapter updates. Catalogue prioritizes search, readable cover titles and compact filters. Authentication uses direct task headings such as Тіркелу. Content density must not shrink touch targets or make long Kazakh text unreadable.

Keep five primary mobile destinations. Review at320/360px and desktop, with populated, empty, failed-network and long-content cases. Check44px product touch targets, keyboard focus/overlap, zoom, dark/light contrast and reduced motion. Record screenshots after the real interface is changed. Browser emulation does not establish low-end Android performance.

## Skill evaluation

The pre-skill read-only probe already proposed useful task-first layouts and real data; it is not evidence that every generic design failure was reproduced. Its proposal did not include measurable viewport/contrast/zoom checks. The follow-up using the skill supplied concrete checks and distinguished a project44px target from WCAG baseline. A counterexample correctly retained an explicitly approved cream/serif/Inter design rather than imposing a universal font/color ban. No probe output is an approved Qmanga design or proof of rendered quality.

Sources and limitations: [.agents/skills/design-product-ui/references/sources.md](../../.agents/skills/design-product-ui/references/sources.md).
