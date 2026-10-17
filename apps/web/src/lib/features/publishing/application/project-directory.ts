import type { PublishingGateway } from './publishing';

export interface AppliedProjectCriteria {
 query: string;
 teamId: string;
}

function failureCode(failure: unknown): string {
 if (typeof failure === 'object' && failure !== null && 'code' in failure && typeof (failure as { code: unknown }).code === 'string') {
  return (failure as { code: string }).code;
 }
 return 'unavailable';
}

export class ProjectDirectory {
 items: { id: string; titleName: string; teamName: string; teamId?: string }[] = [];
 cursor = '';
 hasMore = false;
 total: number | undefined = undefined;
 loading = false;
 error: string | null = null;
 appliedCriteria: AppliedProjectCriteria = { query: '', teamId: '' };
 private generation = 0;

 constructor(private gateway: Pick<PublishingGateway, 'projectPage'>, private changed: () => void = () => {}) {}

 async search(draftQuery: string, teamId: string) {
  this.generation++;
  this.appliedCriteria = { query: draftQuery.trim(), teamId };
  this.items = [];
  this.cursor = '';
  this.hasMore = false;
  this.total = undefined;
  this.error = null;
  await this.fetchPage(false);
 }

 async more() {
  if (this.loading || !this.hasMore || !this.cursor) return;
  this.error = null;
  await this.fetchPage(true);
 }

 async retry() {
  if (this.loading) return;
  this.error = null;
  if (this.items.length > 0 && this.cursor && this.hasMore) {
   await this.fetchPage(true);
  } else {
   await this.fetchPage(false);
  }
 }

 private async fetchPage(append: boolean) {
  const generation = this.generation;
  this.loading = true;
  this.changed();
  try {
   const batch = await this.gateway.projectPage(this.appliedCriteria, append ? this.cursor : undefined);
   if (generation !== this.generation) return;
   this.items = append
    ? [...new Map([...this.items, ...batch.items].map(item => [item.id, item])).values()]
    : batch.items;
   this.cursor = batch.nextCursor;
   this.hasMore = !!batch.nextCursor;
   this.total = batch.total;
   this.error = null;
  } catch (failure) {
   if (generation !== this.generation) return;
   this.error = failureCode(failure);
  } finally {
   if (generation === this.generation) {
    this.loading = false;
    this.changed();
   }
  }
 }
}
