// https://github.com/vuejs/vue/issues/5298

declare module '*.vue' {
  import Vue from 'vue'
  // eslint-disable-next-line import/no-default-export
  export default Vue
}
