import { Socket } from './Socket';
import { Reference } from './Reference';
import { ClientOperationDelete, ClientOperationSet } from './ClientData';

export const LiquidDb = ({ webSocket }: { webSocket: typeof WebSocket }) => {
    interface DbSettings {
        address?: string;
    }

    class LiquidDb {
        private socket: Socket;

        constructor(
            private settings: DbSettings = {
                address: 'ws://localhost:8080/db'
            }
        ) {}

        initialize(): Promise<any> {
            this.socket = new Socket(this.settings.address, webSocket);

            return new Promise(resolve => {
                if (this.socket.ready) {
                    return resolve(this);
                }

                this.socket.once('ready', () => resolve(this));
            });
        }

        ref(path: string | string[]): Reference {
            return new Reference(path, this.socket);
        }

        //these methods might return useful promises in the future
        delete(path: string[]): Promise<any> {
            this.socket.send({
                operation: ClientOperationDelete,
                path: []
            });

            return Promise.resolve();
        }

        set(data: any): Promise<any> {
            this.socket.send({
                operation: ClientOperationSet,
                path: [],
                value: data
            });

            return Promise.resolve();
        }
    }

    return LiquidDb;
};
