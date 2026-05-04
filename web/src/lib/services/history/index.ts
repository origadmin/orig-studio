// History service barrel export - factory + lazy singletons

import type {IHistoryService} from './types';
import {LocalHistoryService} from './local-history-service';
import {RemoteHistoryService} from './remote-history-service';

/** Create a history service based on authentication state */
export function createHistoryService(isAuthenticated: boolean): IHistoryService {
    if (isAuthenticated) {
        return new RemoteHistoryService();
    }
    return new LocalHistoryService();
}

let _localHistoryService: LocalHistoryService | undefined;
let _remoteHistoryService: RemoteHistoryService | undefined;

export function getLocalHistoryService(): LocalHistoryService {
    if (!_localHistoryService) _localHistoryService = new LocalHistoryService();
    return _localHistoryService;
}

export function getRemoteHistoryService(): RemoteHistoryService {
    if (!_remoteHistoryService) _remoteHistoryService = new RemoteHistoryService();
    return _remoteHistoryService;
}

export * from './types';
export {LocalHistoryService} from './local-history-service';
export {RemoteHistoryService} from './remote-history-service';
export {HistorySyncManager, historySyncManager} from './sync-manager';
