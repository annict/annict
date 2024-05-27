# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include Localizable
      include Loggable
      include SentryLoadable
      include PageCategorizable

      skip_before_action :verify_authenticity_token
      around_action :set_locale
    end
  end
end
