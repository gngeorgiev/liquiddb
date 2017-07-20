import * as WebSocket from 'html5-websocket';

import { LiquidDb } from './LiquidDb';
LiquidDb.initializeShims({ webSocket: WebSocket });

export * from './LiquidDb';
