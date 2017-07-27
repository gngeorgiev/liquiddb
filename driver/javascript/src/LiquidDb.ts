import { Socket } from './Socket';
import { Reference } from './Reference';
import { ClientOperationDelete, ClientOperationSet } from './ClientData';
import { OperationEventData, EventOperation } from './EventData';
import { configure, LogLevel } from './log';
import { Dependencies } from './Dependencies';

export interface DbSettings {
    address?: string;
}

export class LiquidDb {
    private socket: Socket;

    static dependencies: Dependencies;

    constructor(
        private settings: DbSettings = {
            address: 'ws://localhost:8082/db'
        }
    ) {
        this.socket = new Socket(
            this.settings.address,
            LiquidDb.dependencies.webSocket
        );
    }

    static configureLogger(conf: { level: LogLevel }) {
        configure(conf);
    }

    static LogLevel: typeof LogLevel = LogLevel;

    async connect(): Promise<LiquidDb> {
        this.reconnect();

        await new Promise(resolve => {
            if (this.socket.ready()) {
                return resolve(this);
            }

            this.socket.once('ready', () => resolve(this));
        });

        return this;
    }

    get ready() {
        return this.socket.ready;
    }

    reconnect() {
        return this.socket.reconnect();
    }

    close() {
        return this.socket.close();
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

    delete(path: string[]): Promise<OperationEventData> {
        return new Reference([], this.socket).delete();
    }

    set(data: any): Promise<OperationEventData> {
        return new Reference([], this.socket).set(data);
    }

    value() {
        return new Reference([], this.socket).value();
    }

    data(callback: (data: OperationEventData) => any) {
        return new Reference([], this.socket).data(callback);
    }

    on(op: EventOperation, callback: (data: OperationEventData) => any) {
        return new Reference([], this.socket).on(op, callback);
    }

    once(op: EventOperation, callback: (data: OperationEventData) => any) {
        return new Reference([], this.socket).once(op, callback);
    }
}
