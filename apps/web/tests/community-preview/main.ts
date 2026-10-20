import {mount} from 'svelte';
import '../../src/app.css';
import DiscussionFixture from './Discussion.fixture.svelte';
const target=document.getElementById('preview');
const params=new URLSearchParams(location.search);
document.documentElement.dataset.theme=params.get('theme')==='light'?'light':'dark';
document.documentElement.lang=params.get('locale')==='ru'?'ru':'kk';
if(target)mount(DiscussionFixture,{target});
