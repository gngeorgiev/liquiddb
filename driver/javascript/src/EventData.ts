export const EventOperationInsert = 'insert';
export const EventOperationDelete = 'delete';
export const EventOperationUpdate = 'update';
export const EventOperationGet = 'get';
export type EventOperation = 'insert' | 'delete' | 'update' | 'get';

export interface EventData {
    operation: EventOperation;
    path: string[];
    key: string;
    value: any;
}
