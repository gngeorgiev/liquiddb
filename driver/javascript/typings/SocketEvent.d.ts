import { EventOperation, OperationEventData } from './EventData';
export interface SocketEvent {
    event: string;
    path: string[];
    operation: EventOperation;
    callback: (data: OperationEventData) => any;
    id: number;
}
