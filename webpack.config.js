const glob = require('glob');
const path = require('path');

const { WebpackManifestPlugin } = require('webpack-manifest-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');

const isProd = process.env.NODE_ENV === 'production';

const packs = path.join(__dirname, 'app', 'frontend', 'packs');
const targets = glob.sync(path.join(packs, '**/*.js'));
const entry = targets.reduce((entry, target) => {
  const bundle = path.relative(packs, target);
  const ext = path.extname(bundle);

  return Object.assign({}, entry, {
    // Input: "application.js"
    // Output: { "application": "./application.js" }
    [bundle.replace(ext, '')]: `./${path.relative(__dirname, packs)}/${bundle}`,
  });
}, {});

module.exports = {
  entry,
  mode: isProd ? 'production' : 'development',
  output: {
    filename: '[name]-[hash].js',
    chunkFilename: '[name].bundle-[hash].js',
    path: path.resolve(__dirname, 'public', 'packs'),
    publicPath: '/packs/',
  },
  module: {
    rules: [
      {
        test: /\.(js|ts)$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'babel-loader',
          },
        ],
      },
      {
        test: require.resolve('jquery'),
        loader: 'expose-loader',
        options: {
          exposes: ['$', 'jQuery'],
        },
      },
      {
        test: /\.scss$/,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              // you can specify a publicPath here
              // by default it use publicPath in webpackOptions.output
              publicPath: path.resolve(__dirname, 'public', 'packs'),
            },
          },
          {
            loader: 'css-loader', // translates CSS into CommonJS
          },
          {
            loader: 'postcss-loader', // Run post css actions
            options: {
              plugins: function () {
                // post css plugins, can be exported to postcss.config.js
                return [require('precss'), require('autoprefixer')];
              },
            },
          },
          {
            loader: 'sass-loader', // compiles Sass to CSS, using Node Sass by default
          },
        ],
      },
      {
        test: /\.(gif|jpg|jpeg|png|svg)$/,
        use: [
          {
            loader: 'file-loader',
            options: {
              name: '[path][name]-[hash].[ext]',
              context: 'app/frontend',
            },
          },
        ],
      },
    ],
  },
  resolve: {
    alias: {
      vue: isProd ? 'vue/dist/vue.min.js' : 'vue/dist/vue.js',
    },
    modules: ['node_modules', path.resolve(__dirname, 'app', 'frontend')],
    extensions: ['.css', '.gif', '.jpeg', '.jpg', '.js', '.json', '.png', '.scss', '.svg', '.ts'],
  },
  plugins: [
    new WebpackManifestPlugin({
      fileName: 'manifest.json',
      publicPath: '/packs/',
      writeToFileEmit: true,
    }),
    new MiniCssExtractPlugin({
      // Options similar to the same options in webpackOptions.output
      // both options are optional
      filename: '[name]-[hash].css',
      chunkFilename: '[name].bundle-[hash].css',
    }),
    new ForkTsCheckerWebpackPlugin(),
  ],
  devServer: {
    contentBase: path.resolve(__dirname, 'public', 'packs'),
    host: '0.0.0.0',
    port: 8080,
    sockPort: 3001,
    disableHostCheck: true,
  },
  devtool: 'eval-source-map',
};
