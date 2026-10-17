export type Theme = 'system' | 'dark' | 'light';
export function themeFrom(value: string | null): Theme { return value === 'dark' || value === 'light' ? value : 'system'; }
