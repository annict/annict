const Vue = require("vue");

let Flash = Vue.extend({
  template: "#js-ann-flash",
  data: function() {
    return {
      type: gon.flash.type || "notice",
      body: gon.flash.body || ""
    };
  },
  computed: {
    show: function() {
      return !!this.body;
    },
    alertClass: function() {
      switch (this.type) {
        case "notice":
          return "alert-success";
          break;
        case "info":
          return "alert-info";
          break;
        case "alert":
          return "alert-warning";
          break;
        case "danger":
          return "alert-danger";
          break;
      }
    },
    alertIcon: function() {
      switch (this.type) {
        case "notice":
          return "fa-check-circle";
          break;
        case "info":
          return "fa-info-circle";
          break;
        case "alert":
          return "fa-exclamation-circle";
          break;
        case "danger":
          return "fa-exclamation-triangle";
          break;
      }
    }
  },
  methods: {
    close: function() {
      this.body = "";
    }
  },
  ready: function() {
    if (this.show) {
      let self = this;
      setTimeout(function() {
        self.close();
      }, 6000);
    }
  }
});

module.exports = Flash;
