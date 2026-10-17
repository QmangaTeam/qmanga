import { HttpAuthGateway } from '$lib/features/identity/infrastructure/http-auth-gateway';
import type { AuthGateway } from '$lib/features/identity/application/auth';
import type { TeamsGateway } from '$lib/features/teams/application/teams';
import { HttpTeamsGateway } from '$lib/features/teams/infrastructure/http-teams-gateway';
import { ApiClient } from '$lib/shared/infrastructure/api-client';
import type { ThemeStore } from '$lib/features/preferences/application/theme';
import { BrowserThemeStore } from '$lib/features/preferences/infrastructure/browser-theme-store';
export function createPreferences(): ThemeStore { return new BrowserThemeStore(); }
export function createTeams(): TeamsGateway { return new HttpTeamsGateway(new ApiClient(globalThis.fetch.bind(globalThis))); }
export function createAuth(): AuthGateway { return new HttpAuthGateway('', globalThis.fetch.bind(globalThis)); }
