import type { LibraryEntry, ReadingGateway } from './progress';

export interface AppliedLibraryCriteria {
 query: string;
 status: string;
 favorite: boolean;
}

function failureCode(failure: unknown): string {
 if (typeof failure === 'object' && failure !== null && 'code' in failure && typeof (failure as { code: unknown }).code === 'string') {
  return (failure as { code: string }).code;
 }
 return 'unavailable';
}

export class LibraryDirectory {
 items: LibraryEntry[] = [];
 cursor = '';
 hasMore = false;
 total: number | undefined = undefined;
 partial = false;
 busy = false;
 error: string | null = null;
 appliedCriteria: AppliedLibraryCriteria = { query: '', status: '', favorite: false };
 private generation = 0;
 private failedAfter: string | null = null;

 constructor(private gateway: Pick<ReadingGateway, 'list'>, private changed: () => void = () => {}) {}

 clear() {
  this.generation++;
  this.items = [];
  this.cursor = '';
  this.hasMore = false;
  this.total = undefined;
  this.partial = false;
  this.busy = false;
  this.error = null;
  this.failedAfter = null;
  this.changed();
 }

 private matchesSelection(item: LibraryEntry): boolean {
  if (item.removed) return false;
  if (this.appliedCriteria.favorite && !item.favorite) return false;
  if (this.appliedCriteria.status && item.status !== this.appliedCriteria.status) return false;
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
  this.error = null;
  await this.fetchPage(account, '');
 }

 async more(account: string) {
  if (this.busy || !this.hasMore || !this.cursor) return;
  this.error = null;
  await this.fetchPage(account, this.cursor);
 }

 async refresh(account: string) {
  this.generation++;
  this.error = null;
  this.total = undefined;
  this.partial = true;
  await this.fetchPage(account, '');
 }

 async retry(account: string) {
  if (this.busy || this.failedAfter === null) return;
  const after = this.failedAfter;
  this.error = null;
  await this.fetchPage(account, after);
 }

 applyPendingUpdate(item: LibraryEntry) {
  this.partial = true;
  this.total = undefined;
  // The server can match original names and aliases absent from LibraryEntry.
  // Only update confirmed rows; never invent textual matches from a slug.
  if (!this.matchesSelection(item)) {
   this.items = this.items.filter(entry => entry.titleId !== item.titleId);
  } else {
   const existing = this.items.find(entry => entry.titleId === item.titleId);
   if (existing) {
    Object.assign(existing, item);
   }
  }
  this.changed();
 }

 private async fetchPage(account: string, after: string) {
  const generation = this.generation;
  this.failedAfter = null;
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
   this.error = null;
  } catch (failure) {
   if (generation !== this.generation) return;
   this.error = failureCode(failure);
   this.failedAfter = after;
  } finally {
   if (generation === this.generation) {
    this.busy = false;
    this.changed();
   }
  }
 }
}
