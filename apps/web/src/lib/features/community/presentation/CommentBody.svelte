<script lang="ts">
 import {commentSegments} from '../domain/comment-markup';
 let {body}:{body:string}=$props();
</script>
<p class="comment-body">{#each commentSegments(body) as segment}{#if segment.kind==='bold'}<strong>{segment.text}</strong>{:else if segment.kind==='italic'}<em>{segment.text}</em>{:else if segment.kind==='strike'}<s>{segment.text}</s>{:else}{segment.text}{/if}{/each}</p>
<style>.comment-body{white-space:pre-wrap;overflow-wrap:anywhere;font-size:14px;line-height:1.6}</style>
