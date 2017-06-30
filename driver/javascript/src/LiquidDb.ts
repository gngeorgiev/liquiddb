import { Socket } from './Socket';
import { Reference } from './Reference';
import { ClientOperationDelete, ClientOperationSet } from './ClientData';
import { EventData, EventOperation } from './EventData';

export interface DbSettings {
    address?: string;
}

export interface Shims {
    webSocket?: typeof WebSocket;
}

export class LiquidDb {
    private socket: Socket;
    private static shims: Shims;

    constructor(
        private settings: DbSettings = {
            address: 'ws://localhost:8080/db'
        }
    ) {}

    static initializeShims(shims: Shims) {
        LiquidDb.shims = shims;
    }

    initialize(): Promise<LiquidDb> {
        this.socket = new Socket(
            this.settings.address,
            LiquidDb.shims.webSocket
        );

        return new Promise(resolve => {
            if (this.socket.ready) {
                return resolve(this);
            }

            this.socket.once('ready', () => resolve(this));
        });
    }

    close() {
        this.socket.close();
    }

    ref(path: string | string[]): Reference {
        if (!path) {
            throw new Error(
                'Invalid ref path, must be in the format "foo.bar" or ["foo", "bar"].'
            );
        }

        if (typeof path === 'string') {
            path = path.split('.');
        }

        //we should not create a reference with empty path since it can delete the whole tree
        if (!path.length) {
            throw new Error('Invalid ref path, must have at least one level.');
        }

        return new Reference(path, this.socket);
    }

    delete(path: string[]): Promise<EventData> {
        return new Reference([], this.socket).delete();
    }

    set(data: any): Promise<EventData> {
        return new Reference([], this.socket).set(data);
    }

    value(): Promise<any> {
        return new Reference([], this.socket).value();
    }

    data(callback: (data: EventData) => any): () => any {
        return new Reference([], this.socket).data(callback);
    }

    on(op: EventOperation, callback: (data: EventData) => any): () => any {
        return new Reference([], this.socket).on(op, callback);
    }

    once(op: EventOperation, callback: (data: EventData) => any) {
        return new Reference([], this.socket).once(op, callback);
    }
}
