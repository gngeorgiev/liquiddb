const assert = require('assert');
const LiquidDb = require('../dist/node/index.node');

describe('basic', () => {
    it('should create with proper settings', () => {
        const db = new LiquidDb({ address: 'test' });

        assert.notEqual(null, db.settings);
        assert.notEqual(undefined, db.settings);
        assert.equal(db.settings.address, 'test');
    });

    it('should initialize properly', async () => {
        const db = await new LiquidDb().initialize();

        assert.equal(db.socket.ready, true);
    });
});

describe('crud', () => {
    let db;
    let ref;

    before(async () => {
        db = await new LiquidDb().initialize();
        ref = db.ref('foo');
    });

    afterEach(async () => {
        await db.delete();
    });

    it('should set data', async () => {
        await ref.set('test');
        const value = await ref.value();

        assert.equal(value, 'test');
    });

    it('should set json and get correctly', async () => {
        await ref.set({
            bar: 5
        });

        const value = await db.ref('foo.bar').value();

        assert.equal(value, 5);
    });

    it('should delete all data', async () => {
        await ref.set({
            bar: 5
        });

        await db.delete();
        const value = await db.ref('foo.bar').value();
        assert.notEqual(value, 5);
    });
});
