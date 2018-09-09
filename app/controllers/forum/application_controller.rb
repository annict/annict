# frozen_string_literal: true

module Forum
  class ApplicationController < ActionController::Base
    include Pundit

    include ControllerCommon
    include ViewerIdentifiable
    include Analyzable
    include LogrageSetting
    include FlashMessage
    include ViewSelector
    include RavenContext
    include PageCategoryMethods
    include PageParamsMethods
    include HeadersForFastly

    layout "application"

    helper_method :gon

    before_action :redirect_if_unexpected_subdomain
    before_action :switch_locale
    before_action :set_search_params
    before_action :store_page_category
    before_action :load_new_user
  end
end
