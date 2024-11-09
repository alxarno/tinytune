const path = require('path');

module.exports = {
    performance: {
        hints: false,
        maxEntrypointSize: 512000,
        maxAssetSize: 512000
    },
    entry: [
        __dirname + '/js/index.js',
        __dirname + '/scss/index.scss'
    ],
    output: {
        filename: 'index.min.js',
        path: path.resolve(__dirname, 'assets'),
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                exclude: /node_modules/,
                use: [],
            }, {
                test: /\.scss$/,
                exclude: /node_modules/,
                use: [
                    {
                        loader: 'file-loader',
                        options: { name: '[name].min.css'}
                    },
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: true,
                            sassOptions: {
                                quietDeps: true,
                                silenceDeprecations: ["import"],
                                style: `compressed`,
                            },
                        }
                    }
                ]
            }
        ]
    }
};