import truncate from 'lodash/truncate';

import newLine from '../filters/newLine';

export default {
  template: '#t-omitted-synopsis',

  props: {
    text: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      shortenText: truncate(this.text, { length: 100 }),
      canViewFullSynopsis: false,
    };
  },

  methods: {
    format(text) {
      return newLine(text);
    },

    expand() {
      return (this.canViewFullSynopsis = true);
    },
  },

  mounted() {
    return (this.canViewFullSynopsis = this.text.length === this.shortenText.length);
  },
};
