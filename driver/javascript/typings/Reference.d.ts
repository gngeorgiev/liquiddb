import { Socket } from './Socket';
import { OperationEventData, EventOperation } from './EventData';
export declare class Reference {
    private path;
    private socket;
    constructor(path: string[], socket: Socket);
    value(): Promise<any>;
    data(callback: (data: OperationEventData) => any): () => any;
    on(op: EventOperation, callback: (data: OperationEventData) => any): () => any;
    once(op: EventOperation, callback: (data: OperationEventData) => any): void;
    set(value: any): Promise<OperationEventData>;
    delete(): Promise<OperationEventData>;
}
