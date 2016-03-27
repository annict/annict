module.exports = (val) ->
  return "" unless val

  val.
    replace(/(\n\r){3,}/g, "<br><br>").
    replace(/\n/g, "<br>")
