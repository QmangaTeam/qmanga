export interface FacetOption { value: string; kk: string; ru: string }

// The publication document stores the canonical Russian name, matching the
// vocabulary of the reference catalogue. The panel shows a localized label.
export const countryFacets: FacetOption[] = [
	{ value: 'Корея', kk: 'Корея', ru: 'Корея' },
	{ value: 'Япония', kk: 'Жапония', ru: 'Япония' },
	{ value: 'Китай', kk: 'Қытай', ru: 'Китай' },
	{ value: 'США', kk: 'АҚШ', ru: 'США' },
	{ value: 'Россия', kk: 'Ресей', ru: 'Россия' },
	{ value: 'Франция', kk: 'Франция', ru: 'Франция' }
];

export const sourceFacets: FacetOption[] = [
	{ value: 'Аниме', kk: 'Аниме', ru: 'Аниме' },
	{ value: 'Комикс', kk: 'Комикс', ru: 'Комикс' },
	{ value: 'Додзинси', kk: 'Додзинси', ru: 'Додзинси' },
	{ value: 'Игра', kk: 'Ойын', ru: 'Игра' },
	{ value: 'Манга', kk: 'Манга', ru: 'Манга' },
	{ value: 'Ранобэ', kk: 'Жеңіл роман', ru: 'Ранобэ' },
	{ value: 'Новелла', kk: 'Новелла', ru: 'Новелла' },
	{ value: 'Оригинал', kk: 'Түпнұсқа', ru: 'Оригинал' }
];

export const formatFacets: FacetOption[] = [
	{ value: 'Сборник', kk: 'Жинақ', ru: 'Сборник' },
	{ value: 'Додзинси', kk: 'Додзинси', ru: 'Додзинси' },
	{ value: 'В цвете', kk: 'Түсті', ru: 'В цвете' },
	{ value: 'Сингл', kk: 'Сингл', ru: 'Сингл' },
	{ value: 'Веб', kk: 'Веб', ru: 'Веб' },
	{ value: 'Вебтун', kk: 'Вебтун', ru: 'Вебтун' },
	{ value: 'Ёнкома', kk: 'Ёнкома', ru: 'Ёнкома' },
	{ value: 'Short', kk: 'Қысқа формат', ru: 'Short' }
];

export function facetLabel(options: FacetOption[], value: string, locale: string): string {
	const option = options.find(item => item.value === value);
	if (!option) return value;
	return locale === 'ru' ? option.ru : option.kk;
}
