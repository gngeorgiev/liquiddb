export const ClientOperationSet = 'set';
export const ClientOperationGet = 'get';
export const ClientOperationDelete = 'delete';
export const ClientOperationSubscribe = 'subscribe';
export const ClientOperationUnSubscribe = 'unsubscribe';

export type ClientOperation =
    | 'set'
    | 'get'
    | 'delete'
    | 'subscribe'
    | 'unsubscribe';

export interface ClientData {
    id?: number;
    operation: ClientOperation;
    path?: string[];
    value?: any;
}
