const assert = require('assert');
const { LiquidDb, LiquidDbStats } = require('../');

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
        });

        new LiquidDb().connect().then(d => (db = d));
    });

    it('Should reconnect and get data', done => {
        let db;

        stats.once('data', d => {
            setTimeout(() => db.close() && done(), 100);

            assert.ok(d.connectionsCount > 0);
        });

        stats.close();
        stats.reconnect();

        new LiquidDb().connect().then(d => (db = d));
    });
});
