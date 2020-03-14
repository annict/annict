import _ from 'lodash';

import eventHub from '../eventHub';

const DISPLAY_USERS_LIMIT = 12;

export default {
  template: '#t-work-friends',

  data() {
    return {
      appData: {},
      pageData: {},
      showAll: false,
      usersData: [],
    };
  },

  props: {
    workId: {
      type: Number,
      required: true,
    },
  },

  computed: {
    allUsers() {
      if (!this.usersData || !this.usersData.length) {
        return [];
      }

      const data = this.usersData.filter(ud => {
        return ud.work_id === this.workId;
      })[0];

      if (!data) {
        return [];
      }

      return data.users;
    },

    users() {
      if (this.showAll) {
        return this.allUsers;
      }
      return _.take(this.allUsers, DISPLAY_USERS_LIMIT);
    },

    isMoreUsers() {
      return !this.showAll && this.allUsers.length > DISPLAY_USERS_LIMIT;
    },
  },

  methods: {
    more() {
      return (this.showAll = true);
    },
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$root.appData;
      this.pageData = this.$root.pageData;

      if (!this.appData.isUserSignedIn) {
        return;
      }

      this.usersData = this.pageData.users_data;
    });
  },
};
