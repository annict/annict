export default {
  template: `
    <div class="c-sign-up-modal modal">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-body text-center">
            <p class="display-1 text-muted">
              <i class="fal fa-info-circle"></i>
            </p>
            <p>
              {{ i18n.signUpModal.body }}
            </p>
            <a href="/sign_up" class="btn btn-primary mr-2">
              <i class="fas fa-rocket mr-1"></i>
              {{ i18n.noun.signUp }}
            </a>
            <a href="/sign_in" class="btn btn-outline-secondary">
              {{ i18n.noun.signIn }}
            </a>
          </div>
        </div>
      </div>
    </div>
  `,

  computed: {
    i18n() {
      return AnnConfig.i18n;
    },
  },
};
