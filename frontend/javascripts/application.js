const $ = require("jquery");
const Vue = require("vue");

const Flash = require("./components/flash");

$(() => {
  Vue.component("ann-flash", Flash);

  new Vue({
    el: "#js-annict"
  });
});
