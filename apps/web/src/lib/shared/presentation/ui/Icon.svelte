<script lang="ts">
 import type { IconName } from './types';
 let { name, size = 20 }: { name: IconName; size?: number } = $props();
 const paths: Record<IconName, string> = {
  copy:'M9 9h12v12H9Z M15 9V3H3v12h6',
  link:'M10 13a5 5 0 0 0 7 0l4-4a5 5 0 0 0-7-7l-3 3 M14 11a5 5 0 0 0-7 0l-4 4a5 5 0 0 0 7 7l3-3',
  share:'M8 11l8-5 M8 13l8 5 M8 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0 M22 4a3 3 0 1 1-6 0 3 3 0 0 1 6 0 M22 20a3 3 0 1 1-6 0 3 3 0 0 1 6 0',
  message:'M3 3h18v14H9l-6 4Z M7 7h10 M7 11h7',
  chart:'M4 20V10 M10 20V4 M16 20v-7 M22 20V7',
  play:'m8 4 13 8-13 8Z',
  'eye-off':'M3 3l18 18 M10.6 10.6a2 2 0 0 0 2.8 2.8 M9.8 5.2A11 11 0 0 1 12 5c6 0 10 7 10 7a17 17 0 0 1-3 3.8 M6.4 6.4A20 20 0 0 0 2 12s4 7 10 7a12 12 0 0 0 5.6-1.4',
  smile:'M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0 M8 9h.01 M16 9h.01 M8 14s1 3 4 3 4-3 4-3',
  eye:'M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12 M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0',
  users: 'M3 21v-2a5 5 0 0 1 10 0v2 M16 14a5 5 0 0 1 5 5v2 M11 7a3 3 0 1 1-6 0 3 3 0 0 1 6 0 M16 4a3 3 0 0 1 0 6',
  folder: 'M3 7V4h6l3 3h9v13H3Z',
  file: 'M5 3h9l5 5v13H5Z M14 3v6h5 M8 13h8 M8 17h6',
  shield: 'M12 3 3 7v6c0 4 9 8 9 8s9-4 9-8V7Z m-5 9 3 3 5-5',
  help: 'M9 8a3 3 0 0 1 6 0c0 2-3 2-3 5 M12 17h.01 M22 12a10 10 0 1 1-20 0 10 10 0 0 1 20 0',
  flag: 'M4 21V3 M4 4c5-4 11 4 16 0v10c-5 4-11-4-16 0',
  moon: 'M21 13A9 9 0 0 1 11 3a9 9 0 1 0 10 10Z',
  sun: 'M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0 M12 2v2 M12 20v2 M2 12h2 M20 12h2 M5 5l1.5 1.5 M17.5 17.5 19 19 M5 19l1.5-1.5 M17.5 6.5 19 5',
  monitor: 'M3 3h18v14H3Z M12 17v4 M8 21h8',
  home: 'M3 10 12 3l9 7v10H15v-7H9v7H3Z',
  catalog: 'M3 3h7v7H3z M14 3h7v7h-7z M3 14h7v7H3z M14 14h7v7h-7z',
  bookmark: 'M5 3h14v18l-7-5-7 5Z',
  bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9 M10 21h4',
  menu: 'M4 6h16 M4 12h16 M4 18h16',
  search: 'M21 21l-5-5 M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0',
  close: 'm6 6 12 12 M6 18 18 6',
  left: 'm15 5-7 7 7 7',
  right: 'm9 5 7 7-7 7',
  down: 'm5 9 7 7 7-7',
  up: 'm5 15 7-7 7 7',
  settings: 'M4 7h16 M4 17h16 M8 4v6 M16 14v6',
  check: 'm5 12 4 4L19 6',
  download: 'M12 3v12 m-5-5 5 5 5-5 M4 16v5h16v-5',
  upload: 'M12 16V3 m-5 5 5-5 5 5 M4 16v5h16v-5',
  plus: 'M12 4v16 M4 12h16',
  star: 'm12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-3-5.6 3 1.1-6.2L3 9.6l6.2-.9Z',
  trash: 'M3 6h18 M9 6V3h6v3 M5 6l1 15h12l1-15 M10 10v7 M14 10v7',
  user: 'M20 21v-2a7 7 0 0 0-14 0v2 M17 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0',
  reset: 'M3 12a9 9 0 1 0 2.6-6.4 M3 4v5h5'
 };
</script>

<svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false"><path d={paths[name]}/></svg>

<style>svg { flex: 0 0 auto; }</style>
