# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include PageCategorizable
    include SentryLoadable
    include Loggable
    include Localizable
    include KeywordSearchable

    layout "default"

    helper_method :locale_ja?, :locale_en?, :local_url, :page_category

    before_action :set_sentry_context
    before_action :set_search_params
    around_action :set_locale
  end
end
