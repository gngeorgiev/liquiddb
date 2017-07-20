const dts = require('dts-bundle');

function DtsBundlePlugin() {}
DtsBundlePlugin.prototype.apply = function(compiler) {
    compiler.plugin('done', function() {
        dts.bundle({
            name: libraryName,
            main: 'src/index.d.ts',
            out: '../index.d.ts',
            removeSource: true,
            outputAsModuleFolder: true // to use npm in-package typings
        });
    });
};

module.exports = {
    devtool: 'source-map',
    resolve: {
        extensions: ['.ts', '.js']
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                loader: 'ts-loader'
            }
        ]
    },
    output: {
        library: 'LiquidDb',
        libraryTarget: 'umd',
        sourceMapFilename: 'index.js.map'
    }
};
