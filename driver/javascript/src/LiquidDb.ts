import { Socket } from './Socket';
import { Reference } from './Reference';
import { ClientOperationDelete } from './ClientData';

export const LiquidDb = ({ WebSocket }) => {
    interface DbSettings {
        address?: string;
    }

    class LiquidDb {
        private socket: Socket;

        constructor(
            private settings: DbSettings = {
                address: 'ws://localhost:8080/store'
            }
        ) {}

        initialize(): Promise<any> {
            this.socket = new Socket(this.settings.address, WebSocket);

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

        delete(path: string[]): Promise<any> {
            this.socket.send({
                operation: ClientOperationDelete,
                path: []
            });

            return Promise.resolve();
        }
    }

    return LiquidDb;
};
