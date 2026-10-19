import { test, expect } from 'bun:test';
import Filters from './CatalogueFilters.fixture.svelte';
import { render, elements } from './markup';

test('filter panel follows the reference structure and applies without a submit button', async () => {
	const html = render(Filters, { props: { filter: { genres: ['horror'], kinds: ['manga'], chaptersMin: 20, chaptersMax: 100 } } }).body;
	expect(html).toContain('Тег іздеу');
	expect(html).toContain('Барлық таңдалған тег сәйкес келуі керек');
	expect(html).toContain('Танымал');
	expect(html).toContain('Барлығы');
	expect(html).toContain('Қорқыныш');
	expect(html).toContain('Түрі');
	expect(html).toContain('Тарау саны');
	expect(html).toContain('Шыққан жылы');
	expect(html).toContain('Басылым форматы');
	expect(html).toContain('Түпнұсқа көзі');
	expect(html).toContain('Өндірілген ел');
	expect(html).not.toContain('Қолдану');
	const selected = await elements(html, 'input[type=checkbox][value="horror"][checked]');
	expect(selected.length).toBeGreaterThan(0);
	const kind = await elements(html, 'input[type=checkbox][value="manga"][checked]');
	expect(kind.length).toBeGreaterThan(0);
	const ranges = await elements(html, 'input[type=number]');
	expect(ranges.some(input => input.value === '20')).toBe(true);
	expect(ranges.some(input => input.value === '100')).toBe(true);
});

test('filter panel localizes tags and groups into Russian', () => {
	const html = render(Filters, { props: { language: 'ru' } }).body;
	expect(html).toContain('Поиск тегов');
	expect(html).toContain('Темы');
	expect(html).toContain('Ужасы');
	expect(html).toContain('Возрастной рейтинг');
	expect(html).not.toContain('Қорқыныш');
});

test('popular tag list is limited while the full group stays reachable', () => {
	const html = render(Filters, {}).body;
	const themes = html.slice(html.indexOf('data-group="themes"'), html.indexOf('data-group="elements"'));
	const boxes = themes.match(/type="checkbox"/g) || [];
	// Six theme tags in the fixture, five popular checkboxes plus the full-list entry.
	expect(boxes.length).toBe(5);
	expect(themes).toContain('Барлығы');
});
