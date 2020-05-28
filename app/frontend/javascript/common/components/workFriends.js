import take from 'lodash/take';

import eventHub from '../eventHub';
import lazyLoad from '../../web/utils/lazy-load';

const DISPLAY_USERS_LIMIT = 12;

export default {
  template: `
    <div class="c-work-friends" v-if="users.length">
      <div class="font-weight-bold">
        {{ title }}
      </div>
      <div class="align-items-center px-3 pt-2 row">
        <div class="col col-auto pl-0 pb-2 pr-2" v-for="user in users">
          <a :href="'/@' + user.username">
            <img class="js-lazy rounded-circle" :data-src="user.avatar_url" width="30" height="30" :alt="user.username">
          </a>
        </div>
        <div class="col pb-2 pl-2">
          <div class="u-fake-link" v-if="isMoreUsers" @click="more">
            {{ moreText }}
          </div>
        </div>
      </div>
    </div>
  `,

  data() {
    return {
      showAll: false,
      usersData: [],
    };
  },

  props: {
    workId: {
      type: Number,
      required: true,
    },

    title: {
      type: String,
      required: true,
    },

    moreText: {
      type: String,
      required: true,
    },
  },

  computed: {
    allUsers() {
      if (!this.usersData || !this.usersData.length) {
        return [];
      }

      const data = this.usersData.filter((ud) => {
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
      return take(this.allUsers, DISPLAY_USERS_LIMIT);
    },

    isMoreUsers() {
      return !this.showAll && this.allUsers.length > DISPLAY_USERS_LIMIT;
    },
  },

  methods: {
    more() {
      this.showAll = true;

      this.$nextTick(() => {
        lazyLoad.update();
      });
    },
  },

  mounted() {
    eventHub.$on('request:work-friends:fetched', (result) => {
      this.usersData = result;

      this.$nextTick(() => {
        lazyLoad.update();
      });
    });
  },
};
