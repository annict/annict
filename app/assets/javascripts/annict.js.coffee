window.Annict = {}

namespaces = ['Actions', 'Components', 'Constants', 'Dispatcher', 'Stores', 'Utils']
_.each namespaces, (ns) ->
  Annict[ns] = {}

# Turbolinksのプログレスバーを使用する
# https://github.com/rails/turbolinks/#progress-bar
Turbolinks.enableProgressBar()

# Moment.jsを日本語で使う
# http://momentjs.com/docs/#/i18n/changing-language/
moment.locale('ja')
