import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '#t-username-preview',

  data() {
    return {
      message: gon.I18n['messages.registrations.new.username_preview'],
      username: $('#user_username').val() || '',
    };
  },

  mounted() {
    const self = this;
    return $('#user_username').on('change paste keyup', function() {
      return (self.username = $(this).val());
    });
  },
};
