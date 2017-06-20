import { Socket } from './Socket';
import { EventData, EventOperationGet } from './EventData';
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

        this.path = path;
        this.socket = socket;
    }

    value(): Promise<any> {
        //connect this with the TODO in the server for ids per message
        return new Promise(resolve => {
            const removeListener = this.socket.subscribe(
                this.path,
                EventOperationGet,
                (data: EventData) => {
                    removeListener();
                    resolve(data);
                }
            );

            this.socket.send({
                operation: ClientOperationGet,
                path: this.path
            });
        });
    }

    on(ev: string, callback: (data: EventData) => any): () => any {
        return this.socket.subscribe(this.path, null, callback);
    }

    //TODO: we need to allow chaining set/value, this means that messages should get
    //a meaningful response
    set(value: any): Promise<any> {
        this.socket.send({
            operation: ClientOperationSet,
            path: this.path,
            value: value
        });

        return Promise.resolve(); //TODO: response!!
    }

    delete(): Promise<any> {
        this.socket.send({
            operation: ClientOperationDelete,
            path: this.path
        });

        return Promise.resolve(); //TODO: response!!
    }
}
