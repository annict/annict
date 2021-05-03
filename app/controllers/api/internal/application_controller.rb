# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include Localizable
      include Loggable
      include SentryLoadable
      include PageCategorizable

      helper_method :locale_ja?, :locale_en?, :local_url, :page_category

      before_action :set_sentry_context
      skip_before_action :verify_authenticity_token
    end
  end
end
