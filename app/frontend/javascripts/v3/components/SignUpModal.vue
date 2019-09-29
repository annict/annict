<template>
  <div class="c-sign-up-modal modal">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-body text-center">
          <p class="display-1 text-muted">
            <i class="fas fa-exclamation-triangle"></i>
          </p>
          <p>
            {{ $root.$t('messages._components.signUpModal.body') }}
          </p>
          <a href="/sign_up" class="btn btn-primary">
            <i class="fas fa-rocket mr-1"></i>
            {{ $root.$t('noun.signUp') }}
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
  import { createComponent } from '@vue/composition-api'

  export default createComponent({})
</script>
