CHAR_MAP =
  "&": "&amp;"
  "<": "&lt;"
  ">": "&gt;"

module.exports = (text) ->
  text.replace /[&<>]/g, (char) ->
    CHAR_MAP[char]
