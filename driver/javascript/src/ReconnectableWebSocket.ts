import { EventEmitter } from 'events';
import * as oibackoff from 'oibackoff';
import { logger } from './log';

const log = logger('ReconnectableWebSocket');

export class ReconnectableWebSocket extends EventEmitter {
    private ws: WebSocket;
    private backoff: oibackoff.BackoffInstance;

    private shouldAutoReconnect: boolean = true;
    private socketOpen: boolean = false;
    private isReconnecting: boolean = false;

    reconnect: () => Promise<any>;
    ready(): boolean {
        return this.socketOpen;
    }

    onSocketClose() {}

    onSocketError(ev: Event) {}

    onSocketOpen() {}

    onSocketMessage(ev: Event) {}

    constructor(private address: string, webSocket: typeof WebSocket) {
        super();

        this.backoff = oibackoff.backoff({
            algorithm: 'exponential',
            delayRatio: 0.1,
            maxTries: 0,
            maxDelay: 3
        });

        this.on('close', () => {
            if (this.shouldAutoReconnect && !this.isReconnecting) {
                this.reconnect();
            }
        });

        this.reconnect = () =>
            new Promise(resolve => {
                if (this.ready()) {
                    return resolve();
                }

                if (!this.isReconnecting) {
                    this.isReconnecting = true;

                    let closeListener;
                    let readyListener;

                    this.backoff(
                        (cb: any) => {
                            log.info('Reconnecting...');

                            closeListener = err => cb(err);
                            readyListener = () => {
                                this.isReconnecting = false;
                                this.shouldAutoReconnect = true;

                                this.removeListener('close', closeListener);

                                cb();
                                resolve();
                            };

                            this.once('close', closeListener);
                            this.once('ready', readyListener);

                            this.initWebSocket(webSocket);
                        },
                        err => {
                            if (err) {
                                log.error('Error during reconnection');

                                this.removeListener('close', closeListener);
                                this.removeListener('ready', readyListener);

                                return log.error(err.toString());
                            }
                        },
                        () => {}
                    );
                } else {
                    this.once('ready', () => resolve());
                }
            });
    }

    private initWebSocket(webSocket: typeof WebSocket) {
        this.disposeWebsocket();

        this.ws = new webSocket(this.address);
        this.ws.onclose = (...args) => {
            const err = new Error(`Socket close ${args}`);
            log.error(err);

            this.socketOpen = false;
            this.emit('close', err);
            this.onSocketClose();
        };

        this.ws.onerror = (...args) => {
            log.error(`Socket error ${args}`);
            this.socketOpen = false;

            this.onSocketError(args[0]);
        };

        this.ws.onopen = () => {
            log.info('Socket open');

            this.socketOpen = true;
            this.onSocketOpen();
        };

        this.ws.onmessage = this.onSocketMessage.bind(this);
    }

    disposeWebsocket() {
        this.socketOpen = false;

        if (this.ws) {
            this.ws.onclose = this.ws.onopen = this.ws.onerror = this.ws.onmessage = null;
            this.ws.close();

            this.ws = null;
        }
    }

    close(): Promise<any> {
        this.shouldAutoReconnect = false;

        return new Promise(resolve => {
            this.disposeWebsocket();
            resolve();
        });
    }

    send(data: any) {
        this.ws.send(data);
    }
}
