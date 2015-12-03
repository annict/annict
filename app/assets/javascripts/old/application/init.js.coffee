window.AnnictOld = {}

# Moment.jsを日本語で使う
# http://momentjs.com/docs/#/i18n/changing-language/
moment.locale("ja")

modules = [
  "angulartics"
  "angulartics.google.analytics"
  "ngAnimate"
  "ngSanitize"
  "infinite-scroll"
]
AnnictOld.angular = angular.module("annict", modules)
