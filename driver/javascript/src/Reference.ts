import { Socket } from './Socket';
import {
    EventData,
    EventOperation,
    EventOperationGet,
    EventOperationInsert,
    EventOperationUpdate,
    EventOperationDelete
} from './EventData';
import {
    ClientOperationSet,
    ClientOperationDelete,
    ClientOperationGet
} from './ClientData';

export class Reference {
    private path: string[];
    private socket: Socket;

    constructor(path: string | string[], socket: Socket) {
        if (typeof path === 'string') {
            path = path.split('.');
        }

        if (!path || !path.length) {
            throw new Error('Cannot create a reference with empty path.');
        }

        this.path = path;
        this.socket = socket;
    }

    value(): Promise<any> {
        return new Promise(resolve => {
            const id = this.socket.send({
                operation: ClientOperationGet,
                path: this.path
            });

            const off = this.socket.subscribe(
                this.path,
                EventOperationGet,
                id,
                (data: EventData) => {
                    off();
                    resolve(data.value);
                }
            );
        });
    }

    data(callback: (data: EventData) => any): () => any {
        const offCallbacks = [
            this.on(EventOperationInsert, callback),
            this.on(EventOperationUpdate, callback),
            this.on(EventOperationDelete, callback)
        ];

        return () => offCallbacks.forEach(f => f());
    }

    on(op: EventOperation, callback: (data: EventData) => any): () => any {
        return this.socket.subscribe(this.path, op, 0, callback);
    }

    set(value: any): Promise<EventData> {
        return new Promise(resolve => {
            const id = this.socket.send({
                operation: ClientOperationSet,
                path: this.path,
                value: value
            });

            const off = this.socket.subscribe(
                this.path,
                [EventOperationInsert, EventOperationUpdate],
                id,
                data => {
                    off();
                    resolve(data);
                }
            );
        });
    }

    delete(): Promise<EventData> {
        return new Promise(resolve => {
            const id = this.socket.send({
                operation: ClientOperationDelete,
                path: this.path
            });

            const off = this.socket.subscribe(
                this.path,
                EventOperationDelete,
                id,
                data => {
                    off();
                    resolve(data);
                }
            );
        });
    }
}
