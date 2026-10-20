<script lang="ts">
 import {beforeNavigate} from '$app/navigation';
 import {onMount} from 'svelte';
 import {useLocale,text} from '../locale';
 import {allowDiscard} from '../unsaved';
 let {dirty}:{dirty:boolean}=$props();
 const {locale}=useLocale();
 export function confirmDiscard(){return allowDiscard(dirty,()=>window.confirm(text($locale,'Сақталмаған өзгерістер бар. Оларды жоғалтып, жалғастыру керек пе?','Есть несохранённые изменения. Продолжить с их потерей?')));}
 beforeNavigate(navigation=>{if(dirty&&(navigation.willUnload||!confirmDiscard()))navigation.cancel();});
 onMount(()=>{const warn=(event:BeforeUnloadEvent)=>{if(dirty){event.preventDefault();event.returnValue='';}};window.addEventListener('beforeunload',warn);return()=>window.removeEventListener('beforeunload',warn);});
</script>
