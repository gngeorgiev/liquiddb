const { FuseBox } = require('fuse-box');

const fuse = FuseBox.init({
    sourceMaps: true,
    globals: {
        default: 'SStore'
    },
    homeDir: 'src',
    output: 'dist/$name.js'
});
fuse.bundle('sstore').instructions(`>index.ts`).watch();

fuse.run();
