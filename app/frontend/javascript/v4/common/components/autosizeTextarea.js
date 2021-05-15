import autosize from 'autosize';

export default {
  template: `<textarea @input="handleChange">{{ value }}</textarea>`,

  props: ['value'],

  methods: {
    handleChange(e) {
      this.$emit('input', e.target.value);
    },
  },

  mounted() {
    autosize(this.$el);
  },
};
