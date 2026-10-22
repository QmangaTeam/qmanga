import {error} from '@sveltejs/kit';
import {createPublicCatalogue} from '$lib/server/composition';
import type {PageServerLoad} from './$types';
import type {ChapterFilter} from '$lib/features/catalog/application/catalog';
export const load:PageServerLoad=async({fetch,params,url})=>{
 const gateway=createPublicCatalogue(fetch);
 let title;
 try{title=await gateway.detail(params.slug);}catch{error(503,'unavailable');}
 if(!title)error(404,'not_found');
 const chapterFilter:ChapterFilter={team:url.searchParams.get('team')||'',query:url.searchParams.get('q')||'',sort:url.searchParams.get('sort')==='desc'?'desc':'asc'};
 const cursor=url.searchParams.get('cursor')||undefined;
 let chapters,firstChapter;
 try{
  chapters=await gateway.chapters(params.slug,cursor,chapterFilter);
  firstChapter=cursor||chapterFilter.team||chapterFilter.query||chapterFilter.sort==='desc'?(await gateway.chapters(params.slug,undefined,{team:chapterFilter.team,limit:1})).items[0]:chapters.items[0];
 }catch{error(503,'unavailable');}
 const next=new URLSearchParams(url.searchParams);
 if(chapters.nextCursor)next.set('cursor',chapters.nextCursor);
 return {tab:['about','chapters','comments','stats'].includes(url.searchParams.get('tab')||'')?url.searchParams.get('tab')!:'chapters',title,chapters,firstChapter:firstChapter||null,chapterFilter,nextURL:chapters.nextCursor?'/titles/'+encodeURIComponent(params.slug)+'?'+next+'#chapters':null};
};
