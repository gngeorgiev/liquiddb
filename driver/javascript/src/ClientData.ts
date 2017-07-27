export const ClientOperationSet = 'set';
export const ClientOperationGet = 'get';
export const ClientOperationDelete = 'delete';
export const ClientOperationSubscribe = 'subscribe';
export const ClientOperationUnSubscribe = 'unsubscribe';
export const ClientOperationHearthbeatResponse = 'hearthbeatResponse';

export type ClientOperation =
    | 'set'
    | 'get'
    | 'delete'
    | 'subscribe'
    | 'unsubscribe'
    | 'hearthbeatResponse';

export interface ClientData {
    id?: number;
    timestamp?: string;
    operation: ClientOperation;
    path?: string[];
    value?: any;
}
