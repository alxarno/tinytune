var webpack = require('webpack')
const path = require('path');
module.exports = {
    performance: {
        hints: false,
        maxEntrypointSize: 512000,
        maxAssetSize: 512000
    },
    entry: {
        vendor: ['bootstrap', 'video.js', 'js-cookie', 'fslightbox']
    },
    output: {
        filename: 'vendor.bundle.js',
        path: path.resolve(__dirname, 'assets'),
        library: 'vendor_lib'
    },
    plugins: [
        new webpack.DllPlugin({
            name: 'vendor_lib',
            path: path.join(__dirname, 'assets', 'vendor-manifest.json')
        })
    ]
}