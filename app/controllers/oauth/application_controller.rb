# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include Analyzable
    include Gonable
    include FlashMessage
    include ViewSelector

    helper_method :client_uuid, :gon

    layout "application"

    before_action :set_search_params

    private

    def set_search_params
      @search = SearchService.new(params[:q])
    end
  end
end
