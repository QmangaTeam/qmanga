// Presentation only: actual values stay available while the decorative fill enters.
export function revealFill(node:HTMLElement){
 if(typeof IntersectionObserver==='undefined')return;
 let first=0,second=0,disposed=false;
 node.dataset.reveal='pending';
 const observer=new IntersectionObserver(entries=>{
  if(!entries.some(entry=>entry.isIntersecting))return;
  observer.disconnect();
  first=requestAnimationFrame(()=>{second=requestAnimationFrame(()=>{if(!disposed)node.dataset.reveal='shown';});});
 },{threshold:0.1});
 observer.observe(node);
 return {destroy(){disposed=true;observer.disconnect();cancelAnimationFrame(first);cancelAnimationFrame(second);}};
}
