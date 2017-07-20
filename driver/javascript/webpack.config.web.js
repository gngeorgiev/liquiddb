const path = require('path');
const { merge } = require('lodash');

const config = require('./webpack.config.common');

module.exports = merge({}, config, {
    entry: './src/index.web.ts',
    target: 'web',
    output: {
        filename: 'index.js',
        path: path.join(__dirname, 'web')
    }
});
