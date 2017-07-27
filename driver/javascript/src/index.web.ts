import { LiquidDb, DbSettings } from './LiquidDb';
import { Dependencies } from './Dependencies';
import { LiquidDbStats } from './stats/Stats';

const dependencies = { webSocket: WebSocket };

LiquidDb.dependencies = dependencies;
LiquidDbStats.dependencies = dependencies;

export { LiquidDb, DbSettings, Dependencies, LiquidDbStats };
