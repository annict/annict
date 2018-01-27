const path = require('path')

const _ = require('lodash')

const { environment } = require('@rails/webpacker')
const vue = require('./loaders/vue')

environment.loaders.append('vue', vue)

const sassConfig = environment.loaders.get('sass')
const sassLoaderIndex = _.findIndex(sassConfig.use, u => {
  return u.loader === 'sass-loader'
})
const sassLoaderOptions = sassConfig.use[sassLoaderIndex].options
environment.loaders.get('sass').use[sassLoaderIndex].options = Object.assign(sassLoaderOptions, {
  data: [
    '@import "~bootstrap/scss/functions";',
    '@import "~bootstrap/scss/variables";',
    '@import "common/variables/annict";',
    '@import "common/variables/bootstrap";',
  ].join(' '),
  includePaths: [path.resolve(__dirname, '../../app/packs/stylesheets/')],
})

module.exports = environment
