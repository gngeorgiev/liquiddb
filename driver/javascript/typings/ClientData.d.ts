export declare const ClientOperationSet = "set";
export declare const ClientOperationGet = "get";
export declare const ClientOperationDelete = "delete";
export declare const ClientOperationSubscribe = "subscribe";
export declare const ClientOperationUnSubscribe = "unsubscribe";
export declare type ClientOperation = 'set' | 'get' | 'delete' | 'subscribe' | 'unsubscribe';
export interface ClientData {
    id?: number;
    timestamp?: string;
    operation: ClientOperation;
    path?: string[];
    value?: any;
}
