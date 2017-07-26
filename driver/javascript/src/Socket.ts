import { EventEmitter } from 'events';
import * as MersenneTwister from 'mersenne-twister';
import { utc, Moment } from 'moment';

import { SocketEvent } from './SocketEvent';
import {
    ClientData,
    ClientOperationSubscribe,
    ClientOperationUnSubscribe
} from './ClientData';
import {
    BaseEventData,
    EventOperation,
    OperationEventData,
    HearthbeatEventData,
    EventOperationHearthbeat
} from './EventData';

import { logger } from './log';

const log = logger('Socket');
const javascriptUnixTimeLength = 13;

export class Socket extends EventEmitter {
    private socketOpen: boolean;
    private receivedHearthbeat: boolean;
    private generator: MersenneTwister = new MersenneTwister();
    private events: Map<number, SocketEvent> = new Map();
    private serverTime: Moment;
    private lastLocalTimeUpdate: Moment;
    private ws: WebSocket;
    private disconnectedQueue: ClientData[] = [];
    private shouldAutoReconnect: boolean = true;

    get ready(): boolean {
        return this.socketOpen && this.receivedHearthbeat;
    }

    reconnect: () => Promise<any>;

    constructor(private address: string, websocket: typeof WebSocket) {
        super();
        this.reconnect = () =>
            new Promise(resolve => {
                this.once('ready', () => {
                    this.shouldAutoReconnect = true;
                    resolve();
                });
                log.info('Reconnecting...');
                this.initWebSocket(websocket);
            });

        this.reconnect();
    }

    private initWebSocket(webSocket: typeof WebSocket) {
        this.ws = new webSocket(this.address);
        this.ws.onclose = this.onSocketClose.bind(this);
        this.ws.onerror = this.onSocketError.bind(this);
        this.ws.onopen = this.onSocketOpen.bind(this);
        this.ws.onmessage = this.onSocketMessage.bind(this);
    }

    private onSocketClose() {
        this.emit('close');

        this.socketOpen = false;
        this.receivedHearthbeat = false;
        if (this.shouldAutoReconnect) {
            this.reconnect();
        }
    }

    private onSocketError(error: Error) {
        log.error(error);
        this.onSocketClose();
    }

    private onSocketOpen() {
        this.socketOpen = true;
        log.info('Connected!');

        this.once('ready', () => {
            this.disconnectedQueue.forEach(d => this.send(d));
            this.disconnectedQueue = [];

            for (let event of this.events.values()) {
                this.subscribe(
                    event.path,
                    event.operation,
                    event.id,
                    event.callback
                );
            }
        });
    }

    private onSocketMessage(msg: MessageEvent) {
        const data: BaseEventData = JSON.parse(msg.data);
        if (data.operation === EventOperationHearthbeat) {
            this.processHearthbeatEventData(data as HearthbeatEventData);
        } else {
            this.processOperationEventData(data as OperationEventData);
        }
    }

    private processOperationEventData(data: OperationEventData) {
        //TODO: make sure that subscribers without id should receive id based messages as well
        //so far, I think they should
        log.debug(`-OnOperationSocketMessage-`);
        log.debug(`Event Data: ${JSON.stringify(data)}`);
        log.debug(
            `Events: ${JSON.stringify(Object.keys((<any>this)._events))}`
        );

        const events = [
            this.buildEventPath(data.path, data.operation, 0),
            this.buildEventPath(data.path, data.operation, data.id),
            this.buildEventPath([], data.operation, 0),
            this.buildEventPath([], data.operation, data.id)
        ];

        events.forEach(ev => this.emit(ev, data));
    }

    private processHearthbeatEventData(data: HearthbeatEventData) {
        log.debug(`-OnHearthbeatSocketMessage-`);
        log.debug(`Hearthbeat: ${JSON.stringify(data)}`);

        if (!this.receivedHearthbeat) {
            this.receivedHearthbeat = true;
            this.lastLocalTimeUpdate = utc();

            this.emit('ready');
        }

        this.serverTime = this.lastLocalTimeUpdate = utc(data.timestamp);
    }

    private buildEventPath(path: string[], op: EventOperation, id: number) {
        const parts = [op || null, id ? String(id) : null].filter(a => a);
        return path.concat(parts).join('.');
    }

    private unsubscribeImpl(socketEvent: SocketEvent) {
        const { path, event, operation, id, callback } = socketEvent;

        this.removeListener(event, callback);
        this.events.delete(id);

        this.send({
            id,
            path,
            operation: ClientOperationUnSubscribe,
            value: operation
        });
    }

    private subscribeImp(
        path: string[],
        op: EventOperation,
        callback: (data: OperationEventData) => any,
        id: number
    ): SocketEvent {
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

    private updateServerTimeWithDelta() {
        //we save the server time on each hearthbeat, then on each message we gotta
        //update the serverTime with the delta of the passed time since the last message
        const lastUpdate = this.lastLocalTimeUpdate;
        const currentUpdate = utc();
        const delta = currentUpdate.diff(lastUpdate);

        this.lastLocalTimeUpdate = currentUpdate;

        this.serverTime = this.serverTime.add(delta, 'milliseconds');
    }

    private ensureClientDataFields(data: ClientData) {
        data.id = data.id || this.generator.random_int();
        data.timestamp = data.timestamp || this.serverTime.toISOString();
    }

    close(): Promise<any> {
        return new Promise(async resolve => {
            this.shouldAutoReconnect = false;
            for (let event of this.events.values()) {
                this.unsubscribeImpl(event);
            }

            this.ws.close();
            this.once('close', () => resolve());
        });
    }

    sendWait(
        data: ClientData,
        path: string[],
        operations: EventOperation | EventOperation[]
    ): Promise<OperationEventData> {
        return new Promise(resolve => {
            //we need to save the id, so we are generating it manually
            const id = this.generator.random_int();

            this.subscribeOnce(path, operations, id, data => resolve(data));

            data.id = id;
            this.send(data);
        });
    }

    send(data: ClientData): number {
        if (!this.ready) {
            return this.disconnectedQueue.push(data);
        }

        this.updateServerTimeWithDelta();
        this.ensureClientDataFields(data);

        const d = JSON.stringify(data);
        this.ws.send(d);

        log.debug('-Send message-');
        log.debug(`Message: ${d}`);

        return data.id;
    }

    subscribeOnce(
        path: string[],
        operations: EventOperation | EventOperation[],
        id: number,
        callback: (data: OperationEventData) => any
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
        callback: (data: OperationEventData) => any,
        once?: boolean
    ): () => any {
        if (!Array.isArray(operations)) {
            operations = [operations];
        }

        const events = operations.map(op =>
            this.subscribeImp(path, op, callback, id)
        );

        return () => events.forEach(ev => this.unsubscribeImpl(ev));
    }
}
