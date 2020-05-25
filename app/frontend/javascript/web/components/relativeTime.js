import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

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

  computed: {
    timeAgo() {
      const currentTime = dayjs();
      const time = dayjs(this.time);

      const passageDays = dayjs(currentTime).diff(time, 'day');

      if (passageDays > 3) {
        return time.format('YYYY-MM-DD');
      } else {
        return time.fromNow();
      }
    },

    absoluteTime() {
      return dayjs(this.time).format('YYYY-MM-DD HH:mm');
    },
  },
};
