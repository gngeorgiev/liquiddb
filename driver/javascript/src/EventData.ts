export const EventOperationInsert = 'insert';
export const EventOperationDelete = 'delete';
export const EventOperationUpdate = 'update';
export const EventOperationGet = 'get';
export const EventOperationHearthbeat = 'hearthbeat';

export type EventOperation =
    | 'insert'
    | 'delete'
    | 'update'
    | 'get'
    | 'hearthbeat';

export interface BaseEventData {
    id?: number;
    operation: EventOperation;
}

export interface OperationEventData extends BaseEventData {
    path: string[];
    key: string;
    value: any;
}

export interface HearthbeatEventData extends BaseEventData {
    timestamp: number;
}
