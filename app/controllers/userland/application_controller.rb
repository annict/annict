# frozen_string_literal: true

module Userland
  class ApplicationController < ActionController::Base
    include Pundit

    include ControllerCommon
    include Analyzable
    include LogrageSetting
    include Gonable
    include FlashMessage
    include PageCategoryHelper
    include ViewSelector

    layout "application"

    helper_method :gon

    before_action :redirect_if_unexpected_subdomain
    before_action :switch_languages
    before_action :set_search_params
    before_action :load_data_into_gon
    before_action :load_new_user
  end
end
