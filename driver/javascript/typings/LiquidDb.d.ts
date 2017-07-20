import { Reference } from './Reference';
import { OperationEventData, EventOperation } from './EventData';
import { LogLevel } from './log';
export interface DbSettings {
    address?: string;
}
export interface Shims {
    webSocket?: typeof WebSocket;
}
export declare class LiquidDb {
    private settings;
    private socket;
    private static shims;
    constructor(settings?: DbSettings);
    static initializeShims(shims: Shims): void;
    static configureLogger(conf: {
        level: LogLevel;
    }): void;
    static LogLevel: typeof LogLevel;
    initialize(): Promise<LiquidDb>;
    close(): void;
    ref(path: string | string[]): Reference;
    delete(path: string[]): Promise<OperationEventData>;
    set(data: any): Promise<OperationEventData>;
    value(): Promise<any>;
    data(callback: (data: OperationEventData) => any): () => any;
    on(op: EventOperation, callback: (data: OperationEventData) => any): () => any;
    once(op: EventOperation, callback: (data: OperationEventData) => any): void;
}
