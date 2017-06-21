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

    constructor(path: string[], socket: Socket) {
        this.path = path;
        this.socket = socket;
    }

    async value(): Promise<any> {
        const data = await this.socket.sendWait(
            {
                operation: ClientOperationGet,
                path: this.path
            },
            this.path,
            EventOperationGet
        );
        return data.value;
    }

    data(callback: (data: EventData) => any): () => any {
        const offCallbacks = [
            this.on(EventOperationInsert, callback),
            this.on(EventOperationUpdate, callback),
            this.on(EventOperationDelete, callback)
        ];

        return () => offCallbacks.forEach(f => f());
    }

    //TODO: once etc
    on(op: EventOperation, callback: (data: EventData) => any): () => any {
        return this.socket.subscribe(this.path, op, 0, callback);
    }

    async set(value: any): Promise<EventData> {
        const data = await this.socket.sendWait(
            {
                operation: ClientOperationSet,
                path: this.path,
                value: value
            },
            this.path,
            [EventOperationInsert, EventOperationUpdate]
        );

        return data;
    }

    async delete(): Promise<EventData> {
        const data = await this.socket.sendWait(
            {
                operation: ClientOperationDelete,
                path: this.path
            },
            this.path,
            EventOperationDelete
        );
        return data;
    }
}
