import $ from 'jquery'

export default {
  exec() {
    $('<link>')
      .prop({ rel: 'stylesheet', media: 'all', href: 'https://use.fontawesome.com/releases/v5.0.6/css/all.css' })
      .appendTo('head')
  },
}
