var webpack = require('webpack');
var path = require('path');

var isProd = (process.env.NODE_ENV === 'production');

var BUILD_DIR = path.resolve(__dirname, '../sitedata/built');
var APP_DIR = path.resolve(__dirname, 'src');

// Conditionally return a list of plugins to use based on the current environment.
// Repeat this pattern for any other config key (ie: loaders, etc).
function getPlugins() {
  var plugins = [];

  // Always expose NODE_ENV to webpack, you can now use `process.env.NODE_ENV`
  // inside your code for any environment checks; UglifyJS will automatically
  // drop any unreachable code.
  plugins.push(new webpack.DefinePlugin({
    'process.env': {
      'NODE_ENV': JSON.stringify(process.env.NODE_ENV)
    }
  }));


  // Conditionally add plugins for Production builds.
  if (isProd) {
    // This helps ensure the builds are consistent if source hasn't changed
    plugins.push(new webpack.optimize.OccurrenceOrderPlugin())
    // Try to dedupe duplicated modules, if any
    plugins.push(new webpack.optimize.DedupePlugin())
    // Minify the code.
    plugins.push(new webpack.optimize.UglifyJsPlugin({
      compress: {
        screw_ie8: true, // React doesn't support IE8
        warnings: false
      },
      mangle: {
        screw_ie8: true
      },
      output: {
        comments: false,
        screw_ie8: true
      }
    }));
  }
  return plugins;
}

var config = {
  entry: APP_DIR + '/index.jsx',
  output: {
    path: BUILD_DIR,
    filename: 'bundle.js',
    publicPath: '/built/'
  },
  module: {
    loaders: [
      {
        test: /\.jsx?/,
        include: APP_DIR,
        loader: 'babel',
        query: {
          cacheDirectory: true
        }
      },
      {
        test: /\.css$/,
        loader: "style!css"
      },
    ]
  },
  devtool: "#cheap-module-eval-source-map",
  plugins: getPlugins(),
  resolve: {
    extensions: ['', '.js', '.jsx']
  },
  devServer: {
    inline: true,
    port: 8081,
    proxy: {
      "**": "http://localhost:8080"
    }
  }
};

if (isProd) {
  config.devtool = "#source-map"
  config.stats = {timings: true}
}

module.exports = config;
