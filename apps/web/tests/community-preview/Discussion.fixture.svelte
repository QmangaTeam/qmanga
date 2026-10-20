<script lang="ts">
 import Discussion from '$lib/features/community/presentation/Discussion.svelte';
 import {provideLocale} from '$lib/shared/presentation/locale';
 import type {Comment,CommunityGateway} from '$lib/features/community/application/community';
 const params=new URLSearchParams(location.search);
 provideLocale(params.get('locale')==='ru'?'ru':'kk',()=>{});
 let calls=$state<unknown[]>([]),failed=false,submissionFailed=false;
 const comment=(id:string,userId:string,parentId='',rootId='',body='Талқылаудағы ұзақ пікір. '.repeat(4)):Comment=>({id,userId,parentId,rootId,body,displayName:userId==='own'?'Менің атым':userId==='other'?'ҰзақЕсім'.repeat(9):'Оқырман',spoiler:false,hidden:false,createdAt:'2026-10-20T10:00:00Z',replies:0});
 let comments:Comment[]=[{...comment('root','other'),replies:3},comment('reply','own','root','root','Менің жауабым'),{...comment('nested','third','reply','root','Жасырын жауап'),spoiler:true}];
 const gateway:CommunityGateway={
  account:async()=>params.has('guest')?null:'own',
  discussion:async(_title,_release,after,focus)=>{
   calls=[...calls,{kind:'discussion',after,focus}];
   if(params.has('error')&&!failed){failed=true;throw Error('network_error');}
   return {comments:params.has('empty')?[]:after?[comment('older','third','','','Алдыңғы пікір')]:[...comments],next:after?'':'older-page',score:0,ratings:0};
  },
  replies:async(id,after)=>{calls=[...calls,{kind:'replies',id,after}];return {comments:[comment('remaining','third','root','root','Соңғы жауап')],next:''};},
  comment:async(account,title,release,parent,body,spoiler)=>{if(params.has('submit-error')&&!submissionFailed){submissionFailed=true;throw Error('network_error');}calls=[...calls,{kind:'comment',account,title,release,parent,body,spoiler}];comments=[...comments,{...comment('new','own',parent,parent?'root':'',body),spoiler}];},
  edit:async(account,id,body,spoiler)=>{calls=[...calls,{kind:'edit',account,id,body,spoiler}];comments=comments.map(item=>item.id===id?{...item,body,spoiler}:item);},
  remove:async(account,id)=>{calls=[...calls,{kind:'remove',account,id}];comments=comments.filter(item=>item.id!==id);},
  report:async(account,input)=>{calls=[...calls,{kind:'report',account,...input}];},
  rating:async()=>{},myRating:async()=>0,myReports:async()=>({items:[],hasMore:false}),reports:async()=>({items:[],hasMore:false,next:''}),resolve:async()=>{},visibility:async()=>{}
 };
</script>
<main><p class="fixture-note">Synthetic UI fixture · no account or manga changes</p><Discussion titleId="synthetic-title" titleSlug="synthetic" {gateway}/><output data-testid="calls">{JSON.stringify(calls)}</output></main>
<style>main{max-width:900px;margin:auto;padding:16px}.fixture-note,output{font-size:12px;color:var(--muted)}output{display:block;overflow-wrap:anywhere}</style>
