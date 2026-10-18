export interface MetricDay {date:string;verified:number;publications:number;readers:number;completions:number;ratings:number;libraries:number;subscriptions:number}
export interface MetricReport {days:MetricDay[];startedAt:string;periodReaders:number}
export interface MetricsGateway {report(admin:boolean,days:7|30|90):Promise<MetricReport>}
export function totals(report:MetricReport){return report.days.reduce((sum,day)=>({verified:sum.verified+day.verified,publications:sum.publications+day.publications,completions:sum.completions+day.completions,ratings:sum.ratings+day.ratings,libraries:sum.libraries+day.libraries,subscriptions:sum.subscriptions+day.subscriptions}),{verified:0,publications:0,completions:0,ratings:0,libraries:0,subscriptions:0});}
