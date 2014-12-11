Vue.directive 'ann-loading', (loading) ->
  if loading
    $(@el).append('<div class="loading"><div class="core">Loading...</div></div>')
  else
    $(@el).children('.loading').remove()
