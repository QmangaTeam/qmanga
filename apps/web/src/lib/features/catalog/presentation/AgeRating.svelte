<script lang="ts">
 import {useLocale,text} from '$lib/shared/presentation/locale';
 import type {AgeRating} from '../domain/catalog';
 let {value=0}:{value?:AgeRating}=$props();
 const {locale}=useLocale();
</script>
{#if value>0}<span class="age-rating" class:adult={value===18} aria-label={text($locale,'Жас санаты: ','Возрастная маркировка: ')+value+'+'}>{value}+</span>{/if}
<style>
 .age-rating{display:inline-flex;align-items:center;flex-shrink:0;padding:3px 6px;border-radius:5px;background:var(--raised);color:var(--text);border:1px solid var(--line);font-size:12px;font-weight:750;line-height:1.25;white-space:nowrap}
 .adult{background:var(--age-restricted);color:var(--age-restricted-ink);border-color:var(--age-restricted)}
</style>
