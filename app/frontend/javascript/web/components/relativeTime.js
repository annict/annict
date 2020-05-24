import moment from 'moment';

export default {
  template: `
    <span class='c-relative-time' :title='absoluteTime'>
      {{ timeAgo }}
    </span>
  `,

  props: {
    time: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      datetime: moment(this.time),
    };
  },

  computed: {
    timeAgo() {
      const current = moment();
      const date = this.datetime.format('YYYY-MM-DD');
      const currentDate = current.format('YYYY-MM-DD');

      const passageDays = moment(currentDate).diff(moment(date), 'days');

      if (passageDays > 3) {
        return this.datetime.format('YYYY/MM/DD');
      } else {
        return this.datetime.fromNow();
      }
    },

    absoluteTime() {
      return this.datetime.format('YYYY/MM/DD HH:mm');
    },
  },
};
