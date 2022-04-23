# frozen_string_literal: true

module Userland
  class ApplicationController < ActionController::Base
    include Pundit::Authorization

    include BasicAuthenticatable
    include ControllerCommon
    include PageCategorizable
    include SentryLoadable
    include Loggable
    include Localizable
    include KeywordSearchable
    include ViewSelector

    layout "main_default"

    around_action :set_locale

    private

    def redirect_if_signed_in
      if user_signed_in?
        redirect_to root_path
      end
    end

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
