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
            const interval = setInterval(() => {
                if (db) {
                    clearInterval(interval);
                    db.close();
                    done();
                }
            }, 100);

            assert.ok(d.connections.length > 0);
        });

        new LiquidDb().connect().then(d => {
            db = d;
        });
    });

    it('Should reconnect and get data', done => {
        let db;

        stats.once('data', d => {
            const interval = setInterval(() => {
                if (db) {
                    clearInterval(interval);
                    db.close();
                    done();
                }
            }, 100);

            assert.ok(d.connections.length > 0);
        });

        stats
            .close()
            .then(() => stats.reconnect())
            .then(() => new LiquidDb().connect())
            .then(d => (db = d));
    });
});
