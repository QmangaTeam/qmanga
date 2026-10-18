import {test,expect} from 'bun:test';
import {HttpMetricsGateway} from './http-metrics-gateway';
import {ApiClient} from '$lib/shared/infrastructure/api-client';
import {totals} from '../application/metrics';
function payload(){return {started_at:'2026-10-19T00:00:00Z',period_readers:2,days:Array.from({length:7},(_,index)=>({date:`2026-10-${13+index}`,verified:1,publications:1,readers:2,completions:3,ratings:1,libraries:1,subscriptions:1}))};}
test('maps all event metrics while keeping period readers distinct',async()=>{let path='';const gateway=new HttpMetricsGateway(new ApiClient(async input=>{path=String(input);return Response.json(payload());}));const report=await gateway.report(false,7);expect(path).toBe('/api/v1/studio/metrics?days=7');expect(totals(report).completions).toBe(21);expect(report.periodReaders).toBe(2);expect(report.startedAt).toBe('2026-10-19T00:00:00Z');});
test('rejects malformed or missing counts and periods instead of inventing zero',async()=>{for(const change of [(value:ReturnType<typeof payload>)=>{value.days[0].completions=-1;},(value:ReturnType<typeof payload>)=>{value.days.pop();},(value:ReturnType<typeof payload>)=>{value.period_readers=NaN;}]){const value=payload();change(value);const gateway=new HttpMetricsGateway(new ApiClient(async()=>Response.json(value)));await expect(gateway.report(true,7)).rejects.toThrow();}});
