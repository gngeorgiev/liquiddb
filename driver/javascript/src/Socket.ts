import { EventEmitter } from 'events';
import { EventData, EventOperation } from './EventData';
import { ClientData } from './ClientData';

export class Socket extends EventEmitter {
    private isReady: boolean;

    private ws: WebSocket;

    get ready(): boolean {
        return this.isReady;
    }

    constructor(private address: string) {
        super();

        this.initWebSocket();
    }

    private initWebSocket() {
        this.isReady = false;

        this.ws = new WebSocket(this.address);
        this.ws.onclose = this.onSocketClose.bind(this);
        this.ws.onerror = this.onSocketError.bind(this);
        this.ws.onopen = this.onSocketOpen.bind(this);
        this.ws.onmessage = this.onSocketMessage.bind(this);
    }

    private onSocketClose() {}

    private onSocketError() {}

    private onSocketOpen() {
        this.isReady = true;
        this.emit('ready');
    }

    private onSocketMessage(msg: MessageEvent) {
        this.emit('message', JSON.parse(msg.data));
    }

    private reconnect() {}

    send(data: ClientData): this {
        //TODO: ugly! I think we should change the store api
        const obj = {};
        var currentObj = obj;
        data.path.slice(0, data.path.length - 1).forEach(p => {
            currentObj[p] = {};
            currentObj = currentObj[p];
        });

        currentObj[data.path[data.path.length - 1]] = data.value;
        data.value = obj;

        this.ws.send(JSON.stringify(data));

        return this;
    }

    subscribe(
        path: string[],
        operation: EventOperation,
        callback: (data: EventData) => any
    ): () => any {
        const messageCallback = (data: EventData) => {
            const isSamePath = path.every((el, i) => el === data.path[i]);
            if (!isSamePath) {
                return;
            }

            if (operation === null) {
                return callback(data);
            } else if (operation === data.operation) {
                return callback(data);
            }
        };

        this.on('message', messageCallback);

        return () => this.removeListener('message', messageCallback);
    }
}
