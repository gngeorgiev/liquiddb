import { EventEmitter } from 'events';
import * as MersenneTwister from 'mersenne-twister';

import { EventData, EventOperation } from './EventData';
import { SocketEvent } from './SocketEvent';
import {
    ClientData,
    ClientOperationSubscribe,
    ClientOperationUnSubscribe
} from './ClientData';

export class Socket extends EventEmitter {
    private isReady: boolean;
    private generator: MersenneTwister;
    private events: Map<number, SocketEvent>;

    private ws: WebSocket;

    get ready(): boolean {
        return this.isReady;
    }

    constructor(private address: string, websocket: typeof WebSocket) {
        super();

        this.generator = new MersenneTwister();
        this.events = new Map();

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
            this.buildEventPath(data.path, data.operation, data.id),
            this.buildEventPath([], data.operation, 0),
            this.buildEventPath([], data.operation, data.id)
        ];

        console.log(`
            Event Data: ${msg.data}
            Events: ${JSON.stringify(Object.keys((<any>this)._events))}`);

        events.forEach(ev => this.emit(ev, data));
    }

    private reconnect() {}

    private buildEventPath(path: string[], op: EventOperation, id: number) {
        const parts = [op || null, id ? String(id) : null].filter(a => a);
        return path.concat(parts).join('.');
    }

    private unsubscribeImpl(socketEvent: SocketEvent) {
        const { path, event, operation, id, callback } = socketEvent;

        this.removeListener(event, callback);
        this.send({
            id,
            path,
            operation: ClientOperationUnSubscribe,
            value: operation
        });

        this.events.delete(id);
    }

    private subscribeImp(
        path: string[],
        op: EventOperation,
        callback: (data: EventData) => any,
        id: number
    ): SocketEvent {
        if (!path.length && id) {
        }

        const evPath = this.buildEventPath(path, op, id);
        this.on(evPath, callback);

        this.send({
            path,
            id: id,
            operation: ClientOperationSubscribe,
            value: op
        });

        const event = {
            event: evPath,
            path,
            operation: op,
            callback,
            id
        };

        this.events.set(id, event);

        return event;
    }

    close() {
        this.ws.close();
        for (let [id, event] of this.events.entries()) {
            this.unsubscribeImpl(event);
        }
    }

    sendWait(
        data: ClientData,
        path: string[],
        operations: EventOperation | EventOperation[]
    ): Promise<EventData> {
        return new Promise(resolve => {
            //we need to save the id, so we are generating it manually
            const id = this.generator.random_int();

            this.subscribeOnce(path, operations, id, data => {
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

    subscribeOnce(
        path: string[],
        operations: EventOperation | EventOperation[],
        id: number,
        callback: (data: EventData) => any
    ) {
        const off = this.subscribe(path, operations, id, data => {
            off();
            callback(data);
        });
    }

    subscribe(
        path: string[],
        operations: EventOperation | EventOperation[],
        id: number,
        callback: (data: EventData) => any,
        once?: boolean
    ): () => any {
        if (!Array.isArray(operations)) {
            operations = [operations];
        }

        const events = operations.map(op =>
            this.subscribeImp(path, op, callback, id)
        );

        return () =>
            events.forEach(ev => {
                this.unsubscribeImpl(ev);
            });
    }
}
