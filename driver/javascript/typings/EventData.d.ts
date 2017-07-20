export declare const EventOperationInsert = "insert";
export declare const EventOperationDelete = "delete";
export declare const EventOperationUpdate = "update";
export declare const EventOperationGet = "get";
export declare const EventOperationHearthbeat = "hearthbeat";
export declare type EventOperation = 'insert' | 'delete' | 'update' | 'get' | 'hearthbeat';
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
