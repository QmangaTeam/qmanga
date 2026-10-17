/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	darkMode: 'class',
	theme: {
		extend: {
			screens: {
				xs: '360px'
			},
			minHeight: {
				'touch': '44px'
			},
			minWidth: {
				'touch': '44px'
			}
		}
	},
	plugins: []
};
