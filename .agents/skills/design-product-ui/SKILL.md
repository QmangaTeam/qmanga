---
name: design-product-ui
description: Use when a product interface feels generic, AI-generated, visually noisy, or like a landing-page template, especially during mobile-first UI redesigns. Not for backend architecture or an already approved visual direction that needs faithful implementation.
---

# Design Product UI

Build a recognizable working product, not a demonstration of fashionable CSS. "Less AI" is a contextual design goal, not a measurable certificate or a universal visual style.

## Start with evidence

Read the product brief, existing UI and the user's references. Name the primary task, content, audience, device/network constraints and what the user rejected. Inspect the actual screen; do not infer visual quality from source code.

Write a short visual contract before styling: content priority, density, color roles, typography, navigation and a distinguishing choice justified by the product. Follow the project's approval workflow.

If the current appearance was rejected, compare a materially different direction with the existing one; do not simply keep every token and reduce spacing. If a visual direction is approved, implement it instead of reopening it.

## Design from the task

- Give the first screen to the work: reading, searching, comparing or editing. A catalog needs discoverable content; an account form needs clear fields, not a campaign slogan.
- Establish hierarchy with content grouping, alignment and a restrained type scale. Use cover grids, rows, forms and panels according to their different jobs, not identical cards for everything.
- Define semantic tokens for surfaces, text, actions, selection, borders and feedback. Check real foreground/background combinations. Distinguish status from decoration.
- Choose fonts for legibility, language coverage and delivery cost. Test actual non-Latin text. A system font is valid when justified; changing to an exotic display font is not proof of craftsmanship.
- Let the product's content supply personality. Use concise action labels and relevant metadata. Keep empty, loading, error and offline states distinguishable. Show only real data; mark demonstration content explicitly.

Treat familiar generated motifs as review signals, not universal bans: oversized slogans, highlighted headline fragments, neon-on-black, repeated eyebrow labels, ornamental arrows, floating blobs, excessive rounding, glass and staggered entrances. Keep a motif only when it serves this brief. Replacing them with fashionable cream/serif styling can be the same failure.

## Verify the result

Inspect rendered screens at the project's smallest width, main phone size and desktop. Exercise long labels, empty/populated states, keyboard focus and zoom. Verify text contrast against WCAG, touch targets against the project's requirement, and navigation/keyboard overlap. A project's 44px target is stricter than WCAG 2.2 AA's baseline; do not mislabel the standard.

Keep essential UI available without hover. Respect reduced motion and the network budget. Record actual viewport checks, remaining defects and screenshots. Emulation does not establish physical low-end-device performance.

Keep visual tokens and view components in presentation; preserve application ports and business invariants. Do not introduce architecture or dependencies merely for styling.

Read [sources](references/sources.md) for the evidence and its limits.
