const { LiquidDb } = require('../');

const logLevel = process.env.LIQUID_LOG_LEVEL || 'error';

LiquidDb.configureLogger({
    level: LiquidDb.LogLevel[logLevel]
});
