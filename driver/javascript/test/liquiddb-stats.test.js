const assert = require('assert');
const { LiquidDb, LiquidDbStats } = require('../');

const logLevel = process.env.LIQUID_LOG_LEVEL || 'error';

LiquidDb.configureLogger({
    level: LiquidDb.LogLevel[logLevel]
});

describe('Stats', () => {
    let stats;

    before(async () => {
        stats = await new LiquidDbStats().connect();
    });

    after(() => {
        stats.close();
    });

    it('Should get data on one connection', done => {
        let db;

        stats.once('data', d => {
            setTimeout(() => db.close() && done(), 100);

            assert.ok(d.connectionsCount > 0);
            console.log(d.connectionsCount);
        });

        new LiquidDb().connect().then(d => (db = d));
    });

    it('Should reconnect and get data', done => {
        let db;

        stats.once('data', d => {
            setTimeout(() => db.close() && done(), 100);

            assert.ok(d.connectionsCount > 0);
            console.log(d.connectionsCount);
        });

        stats.close();
        stats.reconnect();

        new LiquidDb().connect().then(d => (db = d));
    });
});
