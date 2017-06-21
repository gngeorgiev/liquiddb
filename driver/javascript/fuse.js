const { FuseBox } = require('fuse-box');

const envs = ['browser', 'node'];

envs.forEach(e => {
    const fuse = FuseBox.init({
        sourceMaps: true,
        globals: {
            default: 'LiquidDb'
        },
        homeDir: 'src',
        output: `dist/$name.${e}.js`
    });

    fuse.bundle('liquiddb').instructions(`>index.${e}.ts`).watch();
    fuse.run();
});
