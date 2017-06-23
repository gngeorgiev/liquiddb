import { EventEmitter } from 'events';
import * as MersenneTwister from 'mersenne-twister';

import { EventData, EventOperation } from './EventData';
import { ClientData } from './ClientData';

export class Socket extends EventEmitter {
    private isReady: boolean;
    private generator: MersenneTwister;
    //there are messages with empty paths, e.g. delete on the whole tree
    //which need to be acknoledged since the response from the server
    //returns the deleted nodes and the event cannot be matched otherwise
    private emptyPathsMap: Map<number, boolean>;

    private ws: WebSocket;

    get ready(): boolean {
        return this.isReady;
    }

    constructor(private address: string, websocket: typeof WebSocket) {
        super();

        this.generator = new MersenneTwister();
        this.emptyPathsMap = new Map();
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

        const events = [
            this.buildEventPath(data.path, data.operation, 0),
            this.buildEventPath(data.path, data.operation, data.id)
        ];
        //some id-based subscriptions have specific paths, that are not the same as the
        //response.
        if (this.emptyPathsMap.has(data.id)) {
            events.push(this.buildEventPath([], data.operation, data.id));
        }

        events.forEach(ev => this.emit(ev, data));
    }

    private reconnect() {}

    private buildEventPath(path: string[], op: EventOperation, id: number) {
        const parts = [op || null, id ? String(id) : null].filter(a => a);
        return ['message'].concat(path).concat(parts).join('.');
    }

    private unsubscribe(
        ev: string,
        callback: (data: EventData) => any,
        id?: number
    ) {
        if (id) {
            this.emptyPathsMap.delete(id);
        }
        this.removeListener(ev, callback);
    }

    close() {
        this.ws.close();
    }

    sendWait(
        data: ClientData,
        path: string[],
        operations: EventOperation | EventOperation[]
    ): Promise<EventData> {
        return new Promise(resolve => {
            const id = this.generator.random_int();

            this.subscribeOnce(path, operations, id, resolve);

            data.id = id;
            this.send(data);
        });
    }

    send(data: ClientData): number {
        data.id = data.id || this.generator.random_int();
        this.ws.send(JSON.stringify(data));
        return data.id;
    }

    subscribeOnce(
        path: string[],
        operations: EventOperation | EventOperation[],
        id: number,
        callback: (data: EventData) => any
    ) {
        const off = this.subscribe(path, operations, id, data => {
            callback(data);
            return off();
        });
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

        //avoid having empty event names
        if (!path.length && operations.length === 1) {
            if (id) {
                this.emptyPathsMap.set(id, true);
            } else {
                throw new Error(
                    'Invalid subscription path. Provide either path or id.'
                );
            }
        }

        const events = operations.map(op => this.buildEventPath(path, op, id));
        events.forEach(ev => this.on(ev, callback));

        return () => events.forEach(ev => this.unsubscribe(ev, callback, id));
    }
}
