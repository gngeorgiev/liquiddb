import { EventEmitter } from 'events';

import { logger } from '../log';
import { StatsData } from './StatsData';
import { Dependencies } from '../Dependencies';
import { ReconnectableWebSocket } from '../ReconnectableWebSocket';

export interface StatsSettings {
    statsAddress?: string;
}

const log = logger('Stats');
export class LiquidDbStats extends ReconnectableWebSocket {
    static dependencies: Dependencies;

    reconnect: () => any;

    constructor(
        statsSettings: StatsSettings = {
            statsAddress: 'ws://localhost:8082/stats'
        }
    ) {
        super(statsSettings.statsAddress, LiquidDbStats.dependencies.webSocket);
    }

    onSocketOpen() {
        this.emit('ready');
    }

    onSocketMessage(msg: MessageEvent) {
        const data: StatsData = JSON.parse(msg.data || '{}');
        this.emit('data', data);
    }

    async connect(): Promise<LiquidDbStats> {
        await this.reconnect();
        return this;
    }
}
