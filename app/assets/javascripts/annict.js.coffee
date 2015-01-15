window.Annict = {}

modules = [
  'angulartics',
  'angulartics.google.analytics',
  'ngAnimate',
  'ngSanitize',
  'infinite-scroll'
]
Annict.angular = angular.module('annict', modules)

# Making it work with CSRF protection
# https://shellycloud.com/blog/2013/10/how-to-integrate-angularjs-with-rails-4
Annict.angular.config ($httpProvider) ->
  authToken = $('meta[name="csrf-token"]').prop('content')
  $httpProvider.defaults.headers.common['X-CSRF-TOKEN'] = authToken

# Moment.jsを日本語で使う
# http://momentjs.com/docs/#/i18n/changing-language/
moment.locale('ja')
