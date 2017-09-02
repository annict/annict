/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
const CHAR_MAP = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;"
};

export default text =>
  text.replace(/[&<>]/g, char => CHAR_MAP[char])
;
