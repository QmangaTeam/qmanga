import { expect, test } from 'bun:test';
import Controls from './Controls.fixture.svelte';
import { elements, render } from './markup';

test('a reusable button cannot accidentally submit its surrounding form', async () => {
 const html = render(Controls).body;
 expect((await elements(html, '[data-case="default"]'))[0]).toMatchObject({ type: 'button', name: 'choice', value: 'save' });
 expect((await elements(html, '[data-case="submit"]'))[0]).toMatchObject({ type: 'submit', form: 'sample-form' });
});

test('busy controls block repeated actions without losing their accessible name', async () => {
 const html = render(Controls).body;
 expect((await elements(html, '[data-case="busy"]'))[0]).toMatchObject({ 'aria-busy': 'true', disabled: '', 'aria-label': 'Сақтау' });
 const disabled = (await elements(html, '[data-case="disabled"]'))[0];
 expect(disabled.disabled).toBe('');
 expect(disabled['aria-busy']).toBeUndefined();
});

test('navigation remains a native link, not a nested interactive control', async () => {
 const html = render(Controls).body;
 expect((await elements(html, '[data-case="link"]'))[0].href).toBe('/catalog');
 expect(await elements(html, 'a button')).toHaveLength(0);
});

test('icon controls have names while their decorative SVGs stay silent', async () => {
 const html = render(Controls).body;
 expect((await elements(html, '[data-case="icon-button"]'))[0]).toMatchObject({ type: 'button', 'aria-label': 'Жабу' });
 const icons = await elements(html, 'svg');
 expect(icons).toHaveLength(2);
 expect(icons.every(icon => icon['aria-hidden'] === 'true' && icon.focusable === 'false')).toBe(true);
});

test('fields associate labels hints and errors with the actual native control', async () => {
 const html = render(Controls).body;
 expect(await elements(html, 'label[for="sample-email"]')).toHaveLength(1);
 expect((await elements(html, 'input#sample-email'))[0]).toMatchObject({ required: '', 'aria-invalid': 'true', 'aria-describedby': 'sample-email-hint sample-email-error' });
 expect(await elements(html, '#sample-email-hint')).toHaveLength(1);
 expect(await elements(html, '#sample-email-error')).toHaveLength(1);
 const select = (await elements(html, 'select#sample-select'))[0];
 expect(select.disabled).toBe('');
 expect(select['aria-describedby']).toBeUndefined();
 expect(select['aria-invalid']).toBeUndefined();
});

test('field text cannot inject markup or invent missing description references', async () => {
 const html = render(Controls).body;
 expect(await elements(html, 'script')).toHaveLength(0);
 expect((await elements(html, 'textarea#sample-note'))[0]['aria-describedby']).toBeUndefined();
 expect(html).toContain('unsafe()');
});
