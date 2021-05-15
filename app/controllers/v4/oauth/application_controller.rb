# frozen_string_literal: true

module V4::Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include V6::PageCategorizable
    include V6::SentryLoadable
    include V6::Loggable
    include V6::Localizable
    include V6::KeywordSearchable

    layout "default"

    helper_method :locale_ja?, :locale_en?, :local_url, :page_category

    before_action :set_sentry_context
    before_action :set_search_params
    around_action :set_locale
  end
end
