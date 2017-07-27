import * as MersenneTwister from 'mersenne-twister';
import { utc, Moment } from 'moment';

import { SocketEvent } from './SocketEvent';
import {
    ClientData,
    ClientOperationSubscribe,
    ClientOperationUnSubscribe,
    ClientOperationGet,
    ClientOperationHearthbeatResponse
} from './ClientData';
import {
    BaseEventData,
    EventOperation,
    OperationEventData,
    HearthbeatEventData,
    EventOperationHearthbeat
} from './EventData';
import { ReconnectableWebSocket } from './ReconnectableWebSocket';

import { logger } from './log';

const log = logger('Socket');
const javascriptUnixTimeLength = 13;

export class Socket extends ReconnectableWebSocket {
    private receivedHearthbeat: boolean;
    private generator: MersenneTwister = new MersenneTwister();
    private events: Map<number, SocketEvent[]> = new Map();
    private serverTime: Moment;
    private lastLocalTimeUpdate: Moment;
    private disconnectedQueue: ClientData[] = [];

    ready(): boolean {
        return super.ready() && this.receivedHearthbeat;
    }

    constructor(address: string, websocket: typeof WebSocket) {
        super(address, websocket);
    }

    onSocketClose() {
        this.receivedHearthbeat = false;
    }

    onSocketOpen() {
        this.once('ready', () => {
            this.disconnectedQueue.forEach(d => this.send(d));
            this.disconnectedQueue = [];

            for (let events of this.events.values()) {
                events.forEach(event => {
                    this.subscribe(
                        event.path,
                        event.operation,
                        event.id,
                        event.callback
                    );
                });
            }
        });
    }

    onSocketMessage(msg: MessageEvent) {
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
        log.verbose(`-OnHearthbeatSocketMessage-`);
        log.verbose(`Hearthbeat: ${JSON.stringify(data)}`);

        this.serverTime = this.lastLocalTimeUpdate = utc(data.timestamp);

        //TODO: should we handle all 3 initial hearthbeats
        //or 1 is fine
        if (!this.ready()) {
            this.receivedHearthbeat = true;
            this.lastLocalTimeUpdate = utc();

            this.emit('ready');
        }

        this.send({ operation: ClientOperationHearthbeatResponse });
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

        //some operations use 0 as id
        if (!this.events.has(id)) {
            this.events.set(id, []);
        }

        this.events.get(id).push(event);

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

    async close(): Promise<any> {
        for (let events of this.events.values()) {
            events.forEach(ev => this.unsubscribeImpl(ev));
        }

        await super.close();
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
        if (!this.ready()) {
            return this.disconnectedQueue.push(data);
        }

        this.updateServerTimeWithDelta();
        this.ensureClientDataFields(data);

        const d = JSON.stringify(data);
        super.send(d);

        if (data.operation !== ClientOperationHearthbeatResponse) {
            log.debug('-Send message-');
            log.debug(`Message: ${d}`);
        } else {
            //lets not spam the console since hearthbeats are pretty frequent
            log.verbose('-Send message-');
            log.verbose(`Message: ${d}`);
        }

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
