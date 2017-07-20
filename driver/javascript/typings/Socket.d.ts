/// <reference types="node" />
import { EventEmitter } from 'events';
import { ClientData } from './ClientData';
import { EventOperation, OperationEventData } from './EventData';
export declare class Socket extends EventEmitter {
    private address;
    private socketOpen;
    private receivedHearthbeat;
    private generator;
    private events;
    private serverTime;
    private lastLocalTimeUpdate;
    private ws;
    readonly ready: boolean;
    constructor(address: string, websocket: typeof WebSocket);
    private initWebSocket(webSocket);
    private onSocketClose();
    private onSocketError(error);
    private onSocketOpen();
    private onSocketMessage(msg);
    private processOperationEventData(data);
    private processHearthbeatEventData(data);
    private reconnect();
    private buildEventPath(path, op, id);
    private unsubscribeImpl(socketEvent);
    private subscribeImp(path, op, callback, id);
    private updateServerTimeWithDelta();
    private ensureClientDataFields(data);
    close(): void;
    sendWait(data: ClientData, path: string[], operations: EventOperation | EventOperation[]): Promise<OperationEventData>;
    send(data: ClientData): number;
    subscribeOnce(path: string[], operations: EventOperation | EventOperation[], id: number, callback: (data: OperationEventData) => any): void;
    subscribe(path: string[], operations: EventOperation | EventOperation[], id: number, callback: (data: OperationEventData) => any, once?: boolean): () => any;
}
