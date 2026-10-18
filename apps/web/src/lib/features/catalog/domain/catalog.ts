export const ageRatings = [0, 12, 16, 18] as const;
export type AgeRating = typeof ageRatings[number];
export function validAgeRating(value: unknown): value is AgeRating { return ageRatings.includes(value as AgeRating); }
export const kinds = ['manga', 'manhwa', 'manhua'] as const;
export type Kind = typeof kinds[number];
export const originalStatuses = ['ongoing', 'completed', 'hiatus', 'cancelled'] as const;
export type OriginalStatus = typeof originalStatuses[number];
export const translationStatuses = ['ongoing', 'completed', 'paused', 'stopped'] as const;
export type TranslationStatus = typeof translationStatuses[number];
export const kindNames: Record<Kind, string> = { manga: 'Манга', manhwa: 'Манхва', manhua: 'Маньхуа' };
export const originalStatusNames: Record<OriginalStatus, string> = { ongoing: 'Жалғасуда', completed: 'Аяқталған', hiatus: 'Үзілісте', cancelled: 'Тоқтатылған' };
export const translationStatusNames: Record<TranslationStatus, string> = { ongoing: 'Аударылуда', completed: 'Толық аударылған', paused: 'Үзілісте', stopped: 'Тоқтатылған' };
