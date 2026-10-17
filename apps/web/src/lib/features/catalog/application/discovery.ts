import type {CatalogGateway,ReleaseUpdate,ResultPage,Summary} from './catalog';
export async function loadRankings(gateway:CatalogGateway,kind=''){
 try{return {available:true,page:await gateway.rankings(kind)};}
 catch{return {available:false,page:{items:[],nextCursor:null} as ResultPage<Summary>};}
}
export async function loadUpdates(gateway:CatalogGateway,after?:string,team?:string){
 try{return {available:true,page:await gateway.updates(after,team)};}
 catch{return {available:false,page:{items:[],nextCursor:null} as ResultPage<ReleaseUpdate>};}
}
export async function loadPicks(gateway:CatalogGateway){
 try{return {available:true,page:await gateway.list({featured:true})};}
 catch{return {available:false,page:{items:[],nextCursor:null} as ResultPage<Summary>};}
}
