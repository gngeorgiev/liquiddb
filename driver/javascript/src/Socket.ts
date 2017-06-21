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
        //TODO: make sure that subscribers without id should receive id based messages as well
        //so far, I think they should

        const data: EventData = JSON.parse(msg.data);

        [
            this.buildEventPath(data.path, data.operation, data.id),
            this.buildEventPath(data.path, data.operation, 0)
        ].forEach(ev => this.emit(ev, data));
    }

    private reconnect() {}

    private buildEventPath(path: string[], op: EventOperation, id: number) {
        const parts = [op || null, id ? String(id) : null].filter(a => a);
        return ['message'].concat(path).concat(parts).join('.');
    }

    sendWait(
        data: ClientData,
        path: string[],
        operations: EventOperation | EventOperation[]
    ): Promise<EventData> {
        return new Promise(resolve => {
            const id = this.generator.random_int();

            const off = this.subscribe(path, operations, id, data => {
                off();
                resolve(data);
            });

            data.id = id;
            this.send(data);
        });
    }

    send(data: ClientData): number {
        data.id = data.id || this.generator.random_int();
        this.ws.send(JSON.stringify(data));
        return data.id;
    }

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
