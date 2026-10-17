import type { Snippet } from 'svelte';
import type { HTMLAnchorAttributes, HTMLButtonAttributes } from 'svelte/elements';

export type ActionVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export interface NavigationItem { href:string; label:string; icon:IconName }
export type PageControl = {label:string; href:string} | {label:string; action:()=>void; busy?:boolean; disabled?:boolean};
export interface ButtonProps extends Omit<HTMLButtonAttributes, 'children'> {
 variant?: ActionVariant;
 busy?: boolean;
 children?: Snippet;
}
export interface ActionLinkProps extends Omit<HTMLAnchorAttributes, 'children' | 'href'> {
 href: string;
 variant?: ActionVariant;
 children?: Snippet;
}
export type IconName = 'home' | 'catalog' | 'bookmark' | 'bell' | 'menu' | 'search' | 'close' | 'left' | 'right' | 'down' | 'up' | 'settings' | 'check' | 'download' | 'upload' | 'plus' | 'star' | 'trash' | 'user' | 'users' | 'folder' | 'file' | 'shield' | 'help' | 'flag' | 'moon' | 'sun' | 'monitor';
export interface FieldControl {
 id: string;
 required: boolean;
 disabled: boolean;
 'aria-describedby'?: string;
 'aria-invalid'?: 'true';
}
