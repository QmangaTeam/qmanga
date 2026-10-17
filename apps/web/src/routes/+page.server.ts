import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

interface HealthData {
	status: string;
	version: string;
	postgres: string;
	s3: string;
}

export const load: PageServerLoad = async ({ fetch }) => {
	const apiUrl = env.API_INTERNAL_URL || 'http://api:8080';

	try {
		const res = await fetch(`${apiUrl}/api/v1/health`);
		if (res.ok) {
			const data: HealthData = await res.json();
			return {
				apiConnected: true,
				health: data
			};
		}
		return {
			apiConnected: false,
			error: `Сервер қатесі: ${res.status}`
		};
	} catch (err) {
		return {
			apiConnected: false,
			error: 'Серверге қосылу мүмкін болмады'
		};
	}
};
