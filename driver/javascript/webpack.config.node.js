const path = require('path');
const { merge } = require('lodash');

const config = require('./webpack.config.common');

module.exports = merge({}, config, {
    entry: './src/index.node.ts',
    target: 'node',
    output: {
        filename: 'index.js',
        path: path.join(__dirname, 'node')
    },
    externals: ['html5-websocket']
});
