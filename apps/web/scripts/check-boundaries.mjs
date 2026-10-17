import { readdirSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import ts from 'typescript';
const failures = [];
function inspect(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const filename = join(directory, entry.name);
    if (entry.isDirectory()) { inspect(filename); continue; }
    if (!filename.endsWith('.ts') || filename.endsWith('.test.ts') || !/[\\/](domain|application)[\\/]/.test(filename)) continue;
    const parsed = ts.createSourceFile(filename, readFileSync(filename, 'utf8'), ts.ScriptTarget.Latest, true);
    function visit(node) {
      if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
        const dependency = node.moduleSpecifier.text;
        const target = dependency.startsWith('.') ? resolve(directory, dependency) : dependency;
        if (/(infrastructure|presentation|composition|server)[\\/]/.test(target) || !dependency.startsWith('.') && !dependency.startsWith('$lib/features/') && !dependency.startsWith('$lib/shared/domain/')) failures.push(filename + ': outward import ' + dependency);
      }
      if (ts.isIdentifier(node) && ['window','document','localStorage','indexedDB','caches','fetch','XMLHttpRequest'].includes(node.text)) failures.push(filename + ': platform reference ' + node.text);
      ts.forEachChild(node, visit);
    }
    visit(parsed);
  }
}
inspect('src/lib');
if (failures.length) { console.error(failures.join('\n')); process.exit(1); }
console.log('Frontend dependency boundaries pass.');
