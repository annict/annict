<template>
  <h1>＼(^o^)／ < {{ greeting }}</h1>
</template>

<script>
  module.exports = {
    data: () => {
      return {
        greeting: 'Hello!'
      }
    }
  }
</script>

<style lang="scss" scoped>
  h1 {
    text-align: center;
  }
</style>
