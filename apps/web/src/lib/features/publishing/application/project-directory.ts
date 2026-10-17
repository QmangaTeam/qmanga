import type { PublishingGateway, ProjectPage, ProjectFilter } from './publishing';

export interface AppliedProjectCriteria {
 query: string;
 teamId: string;
}

export class ProjectDirectory {
 items: { id: string; titleName: string; teamName: string; teamId?: string }[] = [];
 cursor = '';
 hasMore = false;
 total: number | undefined = undefined;
 loading = false;
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
  await this.fetchPage(false);
 }

 async more() {
  if (this.loading || !this.hasMore || !this.cursor) return;
  await this.fetchPage(true);
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
  } finally {
   if (generation === this.generation) {
    this.loading = false;
    this.changed();
   }
  }
 }
}
