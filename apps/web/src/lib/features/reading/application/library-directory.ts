import type { LibraryEntry, LibraryFilter, ReadingGateway } from './progress';

export interface AppliedLibraryCriteria {
 query: string;
 status: string;
 favorite: boolean;
}

export class LibraryDirectory {
 items: LibraryEntry[] = [];
 cursor = '';
 hasMore = false;
 total: number | undefined = undefined;
 partial = false;
 busy = false;
 appliedCriteria: AppliedLibraryCriteria = { query: '', status: '', favorite: false };
 private generation = 0;

 constructor(private gateway: Pick<ReadingGateway, 'list'>, private changed: () => void = () => {}) {}

 matches(item: Pick<LibraryEntry, 'name' | 'status' | 'favorite'>): boolean {
  if (this.appliedCriteria.favorite && !item.favorite) return false;
  if (this.appliedCriteria.status && item.status !== this.appliedCriteria.status) return false;
  if (this.appliedCriteria.query && !item.name.toLowerCase().includes(this.appliedCriteria.query.toLowerCase())) return false;
  return true;
 }

 async search(account: string, draftQuery: string, filter: string) {
  this.generation++;
  this.appliedCriteria = {
   query: draftQuery.trim(),
   status: filter === 'all' || filter === 'favorites' ? '' : filter,
   favorite: filter === 'favorites'
  };
  this.items = [];
  this.cursor = '';
  this.hasMore = false;
  this.total = undefined;
  this.partial = false;
  await this.fetchPage(account, '');
 }

 async more(account: string) {
  if (this.busy || !this.hasMore || !this.cursor) return;
  await this.fetchPage(account, this.cursor);
 }

 async refresh(account: string) {
  this.generation++;
  this.items = [];
  this.cursor = '';
  this.hasMore = false;
  this.total = undefined;
  this.partial = false;
  await this.fetchPage(account, '');
 }

 applyPendingUpdate(item: LibraryEntry) {
  this.partial = true;
  this.total = undefined;
  if (!this.matches(item)) {
   this.items = this.items.filter(entry => entry.titleId !== item.titleId);
  } else {
   const existing = this.items.find(entry => entry.titleId === item.titleId);
   if (existing) Object.assign(existing, item);
  }
  this.changed();
 }

 private async fetchPage(account: string, after: string) {
  const generation = this.generation;
  this.busy = true;
  this.changed();
  try {
   const result = await this.gateway.list(account, after || undefined, this.appliedCriteria);
   if (generation !== this.generation) return;
   this.items = after
    ? [...new Map([...this.items, ...result.items].map(entry => [entry.titleId, entry])).values()]
    : result.items;
   this.hasMore = result.hasMore;
   this.cursor = result.nextCursor || result.items.at(-1)?.titleId || '';
   this.total = result.total;
   this.partial = result.partial === true;
  } finally {
   if (generation === this.generation) {
    this.busy = false;
    this.changed();
   }
  }
 }
}
