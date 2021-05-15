# frozen_string_literal: true

module Forum
  class ApplicationController < ActionController::Base
    include Pundit

    include ControllerCommon
    include Localable
    include Analyzable
    include LogrageSetting
    include Gonable
    include FlashMessage
    include ViewSelector
    include SentryLoadable
    include PageCategorizable
    include V4::UserDataFetchable

    layout "application"

    helper_method :gon, :locale_ja?, :locale_en?, :local_url, :page_category

    around_action :switch_locale
    before_action :set_sentry_context
    before_action :redirect_if_unexpected_subdomain
    before_action :set_search_params
    before_action :store_data_into_gon
    before_action :load_new_user
  end
end
