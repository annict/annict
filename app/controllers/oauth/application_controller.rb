# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include ControllerCommon
    include Analyzable
    include LogrageSetting
    include Gonable
    include FlashMessage
    include ViewSelector
    include RavenContext
    include PageCategoryMethods

    helper_method :gon

    layout "application"

    before_action :redirect_if_unexpected_subdomain
    before_action :switch_locale
    before_action :set_search_params
    before_action :store_data_into_gon
    before_action :store_page_category

    private

    def set_search_params
      @search = SearchService.new(params[:q])
    end
  end
end
