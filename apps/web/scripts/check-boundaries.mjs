import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import ts from 'typescript';
const failures = [];
function inspect(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const filename = join(directory, entry.name);
    if (entry.isDirectory()) { inspect(filename); continue; }
    if (!/\.(ts|svelte)$/.test(filename) || filename.endsWith('.test.ts') || !/[\\/](domain|application|presentation)[\\/]/.test(filename)) continue;
    const presentation = /[\\/]presentation[\\/]/.test(filename);
    const source = readFileSync(filename, 'utf8');
    const content = filename.endsWith('.svelte') ? Array.from(source.matchAll(/<script[^>]*>([\s\S]*?)<\/script>/g), match => match[1]).join('\n') : source;
    const parsed = ts.createSourceFile(filename, content, ts.ScriptTarget.Latest, true);
    function visit(node) {
      if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
        const dependency = node.moduleSpecifier.text;
        const target = dependency.startsWith('.') ? resolve(directory, dependency) : dependency;
        const outward = presentation ? /(infrastructure|composition|server)([\\/]|$)/.test(target) : /(infrastructure|presentation|composition|server)([\\/]|$)/.test(target) || !dependency.startsWith('.') && !dependency.startsWith('$lib/features/') && !dependency.startsWith('$lib/shared/domain/');
        if (outward) failures.push(filename + ': outward import ' + dependency);
      }
      const forbiddenNames = presentation ? ['localStorage','indexedDB','caches','fetch','XMLHttpRequest'] : ['window','document','localStorage','indexedDB','caches','fetch','XMLHttpRequest'];
      if (ts.isIdentifier(node) && forbiddenNames.includes(node.text)) failures.push(filename + ': platform reference ' + node.text);
      ts.forEachChild(node, visit);
    }
    visit(parsed);
  }
}
inspect('src/lib');
if (failures.length) { console.error(failures.join('\n')); process.exit(1); }
console.log('Frontend dependency boundaries pass.');
