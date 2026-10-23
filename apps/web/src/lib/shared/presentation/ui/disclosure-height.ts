export interface DisclosureHeight {expanded:boolean;lines?:number}
// Keep explicit measured endpoints so opening and closing can both interpolate.
export function disclosureHeight(node:HTMLElement,options:DisclosureHeight){
 let current=options,disposed=false;
 const measure=()=>{
  if(disposed)return;
  const style=getComputedStyle(node);
  const collapsed=current.lines?parseFloat(style.lineHeight)*current.lines:parseFloat(style.getPropertyValue('--preview-height'));
  if(!Number.isFinite(collapsed))return;
  node.style.maxHeight=Math.min(node.scrollHeight,current.expanded?node.scrollHeight:collapsed)+'px';
 };
 measure();node.classList.add('disclosure-motion');
 const resize=new ResizeObserver(measure);resize.observe(node);
 const changes=new MutationObserver(measure);changes.observe(node,{subtree:true,childList:true,characterData:true});
 document.fonts.addEventListener('loadingdone',measure);void document.fonts.ready.then(measure);
 return {update(value:DisclosureHeight){current=value;measure();},destroy(){disposed=true;resize.disconnect();changes.disconnect();document.fonts.removeEventListener('loadingdone',measure);node.classList.remove('disclosure-motion');node.style.removeProperty('max-height');}};
}
