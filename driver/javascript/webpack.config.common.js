module.exports = {
    devtool: 'cheap-module-eval-source-map',
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
