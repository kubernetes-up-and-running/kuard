var webpack = require('webpack');
var path = require('path');

var BUILD_DIR = path.resolve(__dirname, '../sitedata/built');
var APP_DIR = path.resolve(__dirname, 'src');

module.exports = {
  entry: './src/index.jsx',
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
      }
    ]
  },
  resolve: {
    extensions: ['*', '.js', '.jsx']
  },
  output: {
    path: BUILD_DIR,
    publicPath: '/built/',
    filename: 'bundle.js'
  },
  plugins: [
    new webpack.HotModuleReplacementPlugin()
  ],
  performance: { hints: false },
  devServer: {
    inline: true,
    hot: true,
    port: 8081,
    proxy: {
      "**": "http://localhost:8080"
    }
  }
};

