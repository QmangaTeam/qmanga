import type { CatalogGateway, Chapter } from '../../catalog/application/catalog';

export class ChapterDirectory {
 items: Chapter[] = [];
 nextCursor: string | null = null;
 busy = false;
 failed = false;
 private generation = 0;
 private query = '';
 private sort: 'asc' | 'desc' = 'desc';

 constructor(private gateway: Pick<CatalogGateway, 'chapters'>, private slug: string, private changed: () => void = () => {}) {}

 async search(query: string, sort: 'asc' | 'desc') {
  this.generation++;
  this.query = query.trim();
  this.sort = sort;
  this.items = [];
  this.nextCursor = null;
  await this.load(false);
 }

 async more() {
  if (this.busy || !this.nextCursor) return;
  await this.load(true);
 }

 private async load(append: boolean) {
  const generation = this.generation;
  this.busy = true;
  this.failed = false;
  this.changed();
  try {
   const page = await this.gateway.chapters(this.slug, append ? this.nextCursor || undefined : undefined, { query: this.query, sort: this.sort, limit: 50 });
   if (generation !== this.generation) return;
   this.items = [...new Map([...(append ? this.items : []), ...page.items].map(item => [item.id, item])).values()];
   this.nextCursor = page.nextCursor;
  } catch {
   if (generation === this.generation) this.failed = true;
  } finally {
   if (generation === this.generation) { this.busy = false; this.changed(); }
  }
 }
}
