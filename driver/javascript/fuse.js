const { FuseBox, Sparky } = require('fuse-box');

const args = process.argv.slice(3);

Sparky.task('build-browser', () => {
    const fuse = FuseBox.init({
        sourceMaps: true,
        globals: {
            default: 'LiquidDb'
        },
        homeDir: 'src',
        output: `dist/browser/$name.js`
    });

    fuse.bundle('liquiddb').instructions(`>index.browser.ts`).watch();
    fuse.run();
});
