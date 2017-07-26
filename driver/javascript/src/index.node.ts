import * as WebSocket from 'html5-websocket';

import { LiquidDb, DbSettings } from './LiquidDb';
import { LiquidDbStats } from './Stats/Stats';

import { Dependencies } from './Dependencies';

const dependencies = { webSocket: WebSocket };

LiquidDb.dependencies = dependencies;
LiquidDbStats.dependencies = dependencies;

export { LiquidDb, DbSettings, Dependencies, LiquidDbStats };
