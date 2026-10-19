import {HttpRightsGateway} from '$lib/features/community/infrastructure/http-rights-gateway';
import {HttpProjectRequestsGateway} from '$lib/features/publishing/infrastructure/http-project-requests-gateway';
export function createProjectRequests(){return new HttpProjectRequestsGateway(new ApiClient(globalThis.fetch.bind(globalThis)));}
import {HttpManagementGateway} from '$lib/features/identity/infrastructure/http-management-gateway';
export function createAccountManagement(){const {api,scope}=scopedApi();return new HttpManagementGateway(api,scope);}
export function createRights(){return new HttpRightsGateway(globalThis.fetch.bind(globalThis));}
import { HttpAuthGateway } from '$lib/features/identity/infrastructure/http-auth-gateway';
import { BrowserLocaleStore } from '$lib/features/preferences/infrastructure/browser-locale-store';
export function createLocalePreferences() { return new BrowserLocaleStore(); }
import { HttpReaderGateway } from '$lib/features/reading/infrastructure/http-reader-gateway';
import { Downloads } from '$lib/features/offline/application/downloads';
import { BrowserDownloadStorage } from '$lib/features/offline/infrastructure/browser-download-storage';
import {SavedImageLoader} from '$lib/features/offline/infrastructure/saved-image-loader';
export function createOfflineReader(){return {...createReader(),...createReading(),images:()=>new SavedImageLoader()};}
import {BrowserProgressStorage} from '$lib/features/reading/infrastructure/browser-progress-storage';
import {HttpReadingGateway} from '$lib/features/reading/infrastructure/http-reading-gateway';
import {HttpCommunityGateway} from '$lib/features/community/infrastructure/http-community-gateway';
import {HttpNotificationsGateway} from '$lib/features/notifications/infrastructure/http-notifications-gateway';
import {HttpAccountScope} from '$lib/features/identity/infrastructure/http-account-scope';
function scopedApi(){const api=new ApiClient(globalThis.fetch.bind(globalThis));return {api,scope:new HttpAccountScope(api)};}
export function createCommunity(){const {api,scope}=scopedApi();return new HttpCommunityGateway(api,scope);}
export function createNotifications(){const {api,scope}=scopedApi();return new HttpNotificationsGateway(api,scope);}
import {LibrarySync,CachedReadingGateway} from '$lib/features/reading/application/library-sync';
import {BrowserLibraryStorage} from '$lib/features/reading/infrastructure/browser-library-storage';
export function createReading(){const {api,scope}=scopedApi();const remote=new HttpReadingGateway(api,scope),librarySync=new LibrarySync(remote,new BrowserLibraryStorage(),()=>crypto.randomUUID());return {readingGateway:new CachedReadingGateway(remote,librarySync),librarySync,progressStorage:new BrowserProgressStorage(),readingChanges};}
export function createDownloads(){return new Downloads(new BrowserDownloadStorage(),()=>Date.now());}
import { HttpCatalogGateway } from '$lib/features/catalog/infrastructure/http-catalog-gateway';
import { BrowserReaderPlatform } from '$lib/features/reading/infrastructure/browser-reader-platform';
import { BrowserImageLoader, BrowserTimer } from '$lib/features/reading/infrastructure/browser-image-loader';
export function createReader() {const api=new ApiClient(globalThis.fetch.bind(globalThis));return {gateway:new HttpReaderGateway(api),catalogue:new HttpCatalogGateway(api),platform:new BrowserReaderPlatform(),images:(first?:HTMLImageElement,cacheable=true)=>new BrowserImageLoader(first,cacheable),timer:new BrowserTimer()};}
import type { AuthGateway } from '$lib/features/identity/application/auth';
import type { TeamsGateway } from '$lib/features/teams/application/teams';
import { HttpTeamsGateway } from '$lib/features/teams/infrastructure/http-teams-gateway';
import { ApiClient } from '$lib/shared/infrastructure/api-client';
import type { ThemeStore } from '$lib/features/preferences/application/theme';
import { BrowserThemeStore } from '$lib/features/preferences/infrastructure/browser-theme-store';
import { HttpCatalogManagementGateway } from '$lib/features/catalog/infrastructure/http-catalog-gateway';
import { HttpPublishingGateway } from '$lib/features/publishing/infrastructure/http-publishing-gateway';
import type { PublishingGateway } from '$lib/features/publishing/application/publishing';
export function createPublishing(): PublishingGateway { return new HttpPublishingGateway(new ApiClient(globalThis.fetch.bind(globalThis))); }
import type { CatalogManagementGateway } from '$lib/features/catalog/application/management';
export function createCatalogueManagement(): CatalogManagementGateway { return new HttpCatalogManagementGateway(new ApiClient(globalThis.fetch.bind(globalThis))); }
export function createPreferences(): ThemeStore { return new BrowserThemeStore(); }
export function createTeams(): TeamsGateway { return new HttpTeamsGateway(new ApiClient(globalThis.fetch.bind(globalThis))); }
export function createAuth(): AuthGateway { return new HttpAuthGateway('', globalThis.fetch.bind(globalThis),'',async()=>{await new BrowserProgressStorage().clearAccounts();await new BrowserLibraryStorage().clear();}); }
import {HttpMetricsGateway} from '$lib/features/metrics/infrastructure/http-metrics-gateway';
export function createMetrics(){return new HttpMetricsGateway(new ApiClient(globalThis.fetch.bind(globalThis)));}
import {HttpAvatarGateway} from '$lib/features/identity/infrastructure/http-avatar-gateway';
export function createAvatars(){return new HttpAvatarGateway(globalThis.fetch.bind(globalThis));}
import {HttpGoogleLinkGateway} from '$lib/features/identity/infrastructure/http-google-link';
export function createGoogleLink(){return new HttpGoogleLinkGateway(new ApiClient(globalThis.fetch.bind(globalThis)));}

import {HttpRegistrationPermit} from '$lib/features/reading/infrastructure/http-registration-permit';
export function createRegistrationImport(){const {api,scope}=scopedApi();return {readingGateway:new HttpReadingGateway(api,scope),progressStorage:new BrowserProgressStorage(),readingChanges,registrationPermit:new HttpRegistrationPermit(api)};}

import {readingChanges} from '$lib/features/reading/infrastructure/reading-changes';

import {HttpTitleChanges} from '$lib/features/catalog/infrastructure/http-title-changes';
export function createTitleChanges(){const {api,scope}=scopedApi();return new HttpTitleChanges(api,scope);}
