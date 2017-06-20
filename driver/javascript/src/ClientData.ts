export const ClientOperationSet = 'set';
export const ClientOperationGet = 'get';
export const ClientOperationDelete = 'delete';
export type ClientOperation = 'set' | 'get' | 'delete';

export interface ClientData {
    operation: ClientOperation;
    path?: string[];
    value?: any;
}
