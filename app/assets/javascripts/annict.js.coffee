window.Annict = {}

modules = [
  'angulartics',
  'angulartics.google.analytics',
  'ngAnimate',
  'ngSanitize',
  'infinite-scroll'
]
Annict.angular = angular.module('annict', modules)

# Moment.jsを日本語で使う
# http://momentjs.com/docs/#/i18n/changing-language/
moment.locale('ja')
