/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
export default function(val) {
  if (!val) { return ""; }

  return val.
    replace(/\n{3,}/g, "<br><br>").
    replace(/\n/g, "<br>");
};
