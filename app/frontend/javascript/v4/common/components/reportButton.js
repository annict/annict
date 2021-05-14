import $ from 'jquery';
import Vue from 'vue';

export default {
  mounted() {
    return $('[data-toggle="tooltip"]').tooltip();
  },
};
