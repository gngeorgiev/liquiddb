import { EventEmitter } from 'events';
import * as MersenneTwister from 'mersenne-twister';

import { EventData, EventOperation } from './EventData';
import { ClientData } from './ClientData';

export class Socket extends EventEmitter {
    private isReady: boolean;
    private generator: MersenneTwister;

    private ws: WebSocket;

    get ready(): boolean {
        return this.isReady;
    }

    constructor(private address: string, websocket: typeof WebSocket) {
        super();

        this.generator = new MersenneTwister();
        this.initWebSocket(websocket);
    }

    private initWebSocket(webSocket: typeof WebSocket) {
        this.isReady = false;

        this.ws = new webSocket(this.address);
        this.ws.onclose = this.onSocketClose.bind(this);
        this.ws.onerror = this.onSocketError.bind(this);
        this.ws.onopen = this.onSocketOpen.bind(this);
        this.ws.onmessage = this.onSocketMessage.bind(this);
    }

    private onSocketClose() {}

    private onSocketError() {}

    private onSocketOpen() {
        this.isReady = true;
        this.emit('ready');
    }

    private onSocketMessage(msg: MessageEvent) {
        const data: EventData = JSON.parse(msg.data);
        const ev = this.buildEventPath(data.path, data.operation, data.id);
        this.emit(ev, data);
    }

    private reconnect() {}

    private buildEventPath(path: string[], op: EventOperation, id: number) {
        const parts = [op || null, id ? String(id) : null].filter(a => a);
        return ['message'].concat(path).concat(parts).join('.');
    }

    send(data: ClientData): number {
        data.id = this.generator.random_int();
        this.ws.send(JSON.stringify(data));
        return data.id;
    }

    //TODO: subscribeOnce, since a lot of methods need to subscribe only once
    subscribe(
        path: string[],
        operations: EventOperation | EventOperation[],
        id: number,
        callback: (data: EventData) => any
    ): () => any {
        if (!Array.isArray(operations)) {
            operations = [operations];
        }

        const events = operations.map(op => {
            const ev = this.buildEventPath(path, op, id);
            this.on(ev, callback);
            return ev;
        });

        return () => events.forEach(ev => this.removeListener(ev, callback));
    }
}
