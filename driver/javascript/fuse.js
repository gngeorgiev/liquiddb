const { FuseBox, Sparky } = require('fuse-box');

const args = process.argv.slice(3);

Sparky.task('build-browser', () => {
    const fuse = FuseBox.init({
        sourceMaps: true,
        globals: {
            default: 'LiquidDb'
        },
        homeDir: 'src',
        output: `browser/$name.js`
    });

    fuse.bundle('index').instructions(`>index.browser.ts`).watch();
    fuse.run();
});
