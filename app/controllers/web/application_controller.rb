# frozen_string_literal: true

module Web
  class ApplicationController < ::ApplicationController
    include Pundit

    include RavenContext
    include Loggable
    include Localizable

    before_action :set_raven_context
    around_action :set_locale

    private

    # Override `Devise::Controllers::Helpers#signed_in_root_path`
    def signed_in_root_path(_resource_or_scope)
      root_path
    end

    # Override `Devise::Controllers::Helpers#after_sign_out_path_for`
    def after_sign_out_path_for(_resource_or_scope)
      root_path
    end
  end
end
