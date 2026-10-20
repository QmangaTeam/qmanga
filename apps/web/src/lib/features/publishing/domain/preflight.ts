import {canSubmit} from './revision';
type Page={id:string;status:string};
export function processingSummary(pages:readonly Page[]){
 const summary={ready:0,processing:0,failed:0,unknown:0,failedIDs:[] as string[]};
 for(const page of pages){
  if(page.status==='ready')summary.ready++;
  else if(page.status==='failed'){summary.failed++;summary.failedIDs.push(page.id);}
  else if(['queued','processing','uploading'].includes(page.status))summary.processing++;
  else summary.unknown++;
 }
 return summary;
}
export type SubmissionObstacle='permission'|'transfer'|'pages'|'failed'|'processing'|'order'|'details';
export function submissionObstacles(revision:{state:string;pages:Page[];source?:string;credits?:string;rightsConfirmed?:boolean},options:{editor:boolean;uploading?:boolean;pending?:boolean;orderDirty?:boolean;detailsDirty?:boolean}):SubmissionObstacle[]{
 const obstacles:SubmissionObstacle[]=[];
 if(!options.editor)obstacles.push('permission');
 if(options.uploading||options.pending)obstacles.push('transfer');
 if(!revision.pages.length)obstacles.push('pages');
 if(revision.pages.some(page=>page.status==='failed'))obstacles.push('failed');
 if(revision.pages.length&&!canSubmit(revision))obstacles.push('processing');
 if(options.orderDirty)obstacles.push('order');
 if(options.detailsDirty||[...(revision.source||'').trim()].length<8||[...(revision.credits||'').trim()].length<2||!revision.rightsConfirmed)obstacles.push('details');
 return obstacles;
}
