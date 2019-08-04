<template>
  <div>
    <ann-navbar></ann-navbar>
    <h1>＼(^o^)／ < {{ greeting }}</h1>
  </div>
</template>

<script>
  import NavBar from '../NavBar.vue'

  export default {
    components: {
      'ann-navbar': NavBar
    },

    data() {
      return {
        greeting: 'Hello!'
      }
    }
  }
</script>
