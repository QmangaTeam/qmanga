export function safeReturnPath(value:string|null|undefined):string {
 if(!value||!value.startsWith('/')||value.startsWith('//')||/[\\\u0000-\u0020]/.test(value))return '/';
 try {
  const url=new URL(value,'https://qmanga.invalid');
  if(url.origin!=='https://qmanga.invalid'||/^\/(api|auth|login|register|forgot-password|reset-password|verify-email)(\/|$)/.test(url.pathname))return '/';
  return url.pathname+url.search+url.hash;
 }catch{return '/';}
}
