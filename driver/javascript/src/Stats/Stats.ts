import { EventEmitter } from 'events';

import { logger } from '../log';
import { StatsData } from './StatsData';
import { Dependencies } from '../Dependencies';

export interface StatsSettings {
    statsAddress?: string;
}

const log = logger('Stats');
export class LiquidDbStats extends EventEmitter {
    private ws: WebSocket;
    private shouldReconnect: boolean = true;

    static dependencies: Dependencies;

    reconnect: () => any;

    constructor(
        private statsSettings: StatsSettings = {
            statsAddress: 'ws://localhost:8082/stats'
        }
    ) {
        super();

        this.reconnect = () =>
            new Promise(resolve => {
                this.initWebsocket(LiquidDbStats.dependencies.webSocket);
                this.once('ready', () => {
                    resolve();
                });
            });
    }

    private initWebsocket(webSocket: typeof WebSocket) {
        this.ws = new webSocket(this.statsSettings.statsAddress);
        this.ws.onclose = this.onSocketClose.bind(this);
        this.ws.onerror = this.onSocketError.bind(this);
        this.ws.onopen = this.onSocketOpen.bind(this);
        this.ws.onmessage = this.onSocketMessage.bind(this);
    }

    private onSocketClose() {
        log.info('Stats socket close');
        this.emit('close');

        if (this.shouldReconnect) {
            this.reconnect();
        }
    }

    private onSocketError(err: Error) {
        log.error(err);
        this.onSocketClose();
    }

    private onSocketOpen() {
        log.info('Stats socket opened');
        this.emit('ready');
    }

    private onSocketMessage(msg: MessageEvent) {
        const data: StatsData = JSON.parse(msg.data || '{}');
        this.emit('data', data);
    }

    async connect(): Promise<LiquidDbStats> {
        this.shouldReconnect = true;
        await this.reconnect();
        return this;
    }

    close() {
        this.shouldReconnect = false;
        this.ws.close();
    }
}
