import {error,redirect} from '@sveltejs/kit';
import type {LayoutServerLoad} from './$types';
export const load:LayoutServerLoad=async({parent,url})=>{
 const {user}=await parent();
 if(!user)redirect(303,'/login?returnTo='+encodeURIComponent(url.pathname+url.search));
 if(user.role!=='admin'||!user.verified)error(403,'Бұл бөлімге рұқсат жоқ / Доступ к разделу закрыт');
 return {};
};
