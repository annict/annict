# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include ControllerCommon
      include Localable
      include Analyzable
      include LogrageSetting
      include RavenContext
      include PageCategoryMethods

      skip_before_action :verify_authenticity_token
      before_action :switch_locale
      before_action :store_page_category
    end
  end
end
