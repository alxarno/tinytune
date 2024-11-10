var webpack = require('webpack')
const path = require('path');

module.exports = {
    optimization: {
        splitChunks: {
            cacheGroups: {
                scssVendor: {
                    test: /scss[\\/]vendor.(s)?css$/,
                    name: "vendor",
                    enforce: true,
                    priority: 30
                },
            }
        }
    },
    performance: {
        hints: false,
        maxEntrypointSize: 512000,
        maxAssetSize: 512000
    },
    entry: [
        __dirname + '/js/index.js',
        __dirname + '/scss/index.scss',
        __dirname + '/scss/vendor.scss',
    ],
    output: {
        filename: 'index.min.js',
        path: path.resolve(__dirname, 'assets'),
    },
    plugins: [
        new webpack.DllReferencePlugin({
            context: __dirname,
            manifest: path.join(__dirname, 'assets', 'vendor-manifest.json')
        }),
    ],
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
                            sourceMap: false,
                            api: "modern-compiler",
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