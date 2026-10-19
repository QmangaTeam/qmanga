<script lang="ts">
 import {onMount,untrack,type Snippet} from 'svelte';
 import type {ReadingGateway,ProgressStorage,ReadingChanges} from '../application/progress';
 import {CoverStatuses} from '../application/cover-statuses';
 let {readingGateway,progressStorage,titleIds,children,readingChanges}:{readingChanges?:ReadingChanges;readingGateway:ReadingGateway;progressStorage:ProgressStorage;titleIds:string[];children:Snippet<[Record<string,string>]>}=$props();
 let version=$state(0),mounted=$state(false);
 const directory=new CoverStatuses(untrack(()=>readingGateway),()=>version++);
 const statuses=$derived.by(()=>{version;return directory.items;});
 $effect(()=>{const ids=titleIds;if(mounted)untrack(()=>{void directory.load(ids);});});
 onMount(()=>{mounted=true;const stop=progressStorage.watchAccountChange(()=>{directory.clear();void directory.load(titleIds);});const stopImports=readingChanges?.subscribe(()=>{void directory.load(titleIds);});return()=>{stop();stopImports?.();directory.close();};});
</script>
{@render children(statuses)}
