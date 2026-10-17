import { mount } from 'svelte';
import '../src/lib/shared/presentation/styles/tokens.css';
import Preview from './Preview.svelte';

const target = document.getElementById('preview');
if (target) mount(Preview, { target });
