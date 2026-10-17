export function publicationReleaseId(link:string,origin:string):string|null{
 try{const target=new URL(link.trim(),origin),parts=target.pathname.split('/');if(target.origin!==origin||target.searchParams.has('preview')||parts.length!==3||parts[1]!=='read'||!/^[a-zA-Z0-9-]{1,36}$/.test(parts[2]))return null;return parts[2];}catch{return null;}
}
