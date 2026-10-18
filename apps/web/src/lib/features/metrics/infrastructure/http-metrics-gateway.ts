import type {MetricsGateway,MetricDay,MetricReport} from '../application/metrics';
import type {ApiClient} from '$lib/shared/infrastructure/api-client';
function object(value:unknown){if(!value||typeof value!=='object')throw new Error('Invalid metrics');return value as Record<string,unknown>;}
function count(value:unknown){if(typeof value!=='number'||!Number.isSafeInteger(value)||value<0)throw new Error('Invalid count');return value;}
export class HttpMetricsGateway implements MetricsGateway {
 constructor(private api:ApiClient){}
 async report(admin:boolean,days:7|30|90):Promise<MetricReport>{const value=object(await this.api.call('/api/v1/'+(admin?'admin':'studio')+'/metrics?days='+days));if(!Array.isArray(value.days)||value.days.length!==days||typeof value.started_at!=='string'||!Number.isFinite(Date.parse(value.started_at)))throw new Error('Invalid report');return {startedAt:value.started_at,periodReaders:count(value.period_readers),days:value.days.map(raw=>{const day=object(raw);if(typeof day.date!=='string'||!/^\d{4}-\d{2}-\d{2}$/.test(day.date))throw new Error('Invalid date');return {date:day.date,verified:count(day.verified),publications:count(day.publications),readers:count(day.readers),completions:count(day.completions),ratings:count(day.ratings),libraries:count(day.libraries),subscriptions:count(day.subscriptions)} satisfies MetricDay;})};}
}
