import { Socket } from './Socket';
import { Reference } from './Reference';

interface StoreSettings {
    address?: string;
}

class SStore {
    private socket: Socket;

    constructor(
        private settings: StoreSettings = {
            address: 'ws://localhost:8080/store'
        }
    ) {}

    initialize(): Promise<any> {
        this.socket = new Socket(this.settings.address);

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
}

export = SStore;
