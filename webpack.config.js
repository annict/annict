const glob = require('glob')
const path = require('path')

const ManifestPlugin = require('webpack-manifest-plugin')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')

const isProd = process.env.NODE_ENV === 'production'

const packs = path.join(__dirname, 'app', 'frontend', 'packs')
const targets = glob.sync(path.join(packs, '**/*.tsx'))
const entry = targets.reduce((entry, target) => {
  const bundle = path.relative(packs, target)
  const ext = path.extname(bundle)

  return Object.assign({}, entry, {
    // Input: "application.js"
    // Output: { "application": "./application.js" }
    [bundle.replace(ext, '')]: `./${path.relative(__dirname, packs)}/${bundle}`,
  })
}, {})

module.exports = {
  entry,
  mode: isProd ? 'production' : 'development',
  // Enable sourcemaps for debugging webpack's output.
  devtool: 'source-map',
  output: {
    filename: '[name]-[hash].js',
    chunkFilename: '[name].bundle-[hash].js',
    path: path.resolve(__dirname, 'public', 'packs'),
    publicPath: '/packs/',
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        loader: 'awesome-typescript-loader',
      },
      {
        enforce: 'pre',
        test: /\.js$/,
        loader: 'source-map-loader',
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
    modules: ['node_modules', path.resolve(__dirname, 'app', 'frontend')],
    extensions: ['.css', '.gif', '.jpeg', '.jpg', '.js', '.json', '.png', '.scss', '.svg', '.ts', '.tsx'],
  },
  plugins: [
    new ManifestPlugin({
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
  ],
  devServer: {
    contentBase: path.resolve(__dirname, 'public', 'packs'),
    host: require('ip').address(),
    port: 8080,
    disableHostCheck: true,
  },
}
