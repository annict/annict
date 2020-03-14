/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
import moment from 'moment';

export default {
  template: `\
<span class='c-time-ago' :title='timeAgoDetail'>
  {{ timeAgo }}
</span>\
`,

  props: {
    time: {
      type: String,
      required: true,
    },
  },

  data() {
    return { datetime: moment(this.time) };
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

    timeAgoDetail() {
      return this.datetime.format('YYYY/MM/DD HH:mm');
    },
  },
};
