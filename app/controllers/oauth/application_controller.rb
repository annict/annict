# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include ControllerCommon
    include Localable
    include Analyzable
    include LogrageSetting
    include Gonable
    include FlashMessage
    include ViewSelector
    include RavenContext
    include V4::UserDataFetchable

    helper_method :gon, :locale_ja?, :locale_en?, :local_url

    layout "application"

    around_action :switch_locale
    before_action :redirect_if_unexpected_subdomain
    before_action :set_search_params
    before_action :store_data_into_gon

    private

    def set_search_params
      @search = SearchService.new(params[:q])
    end
  end
end
