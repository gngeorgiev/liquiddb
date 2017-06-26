import { EventOperation, EventData } from './EventData';
export interface SocketEvent {
    event: string;
    path: string[];
    operation: EventOperation;
    callback: (data: EventData) => any;
    id: number;
}
