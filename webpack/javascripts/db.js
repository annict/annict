const $ = require("jquery");

// Need to load UIkit JS library
window.jQuery = $;

const Turbolinks = require("turbolinks");
const rails = require("jquery-ujs");
const Vue = require("vue");
const UIkit = require("uikit");

$(document).on("turbolinks:load", function() {
  Vue.config.debug = true;
  console.log('turbolinks:load');

  new Vue({
    el: "body"
  });
});

Turbolinks.start();
