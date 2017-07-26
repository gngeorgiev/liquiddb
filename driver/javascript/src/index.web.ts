import { LiquidDb, DbSettings } from './LiquidDb';
import { Dependencies } from './Dependencies';
import { LiquidDbStats } from './Stats/Stats';

const dependencies = { webSocket: WebSocket };

LiquidDb.dependencies = dependencies;
LiquidDbStats.dependencies = dependencies;

export { LiquidDb, DbSettings, Dependencies, LiquidDbStats };
