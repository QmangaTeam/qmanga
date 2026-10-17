export { render } from 'svelte/server';

export async function elements(body: string, selector: string): Promise<Record<string, string>[]> {
 const matches: Record<string, string>[] = [];
 await new HTMLRewriter().on(selector, {
  element(element) { matches.push(Object.fromEntries(element.attributes)); }
 }).transform(new Response(body)).text();
 return matches;
}
