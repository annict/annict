# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include ControllerCommon
      include Localable
      include Analyzable
      include LogrageSetting
      include RavenContext
      include PageCategorizable

      helper_method :locale_ja?, :locale_en?, :local_url, :page_category

      skip_before_action :verify_authenticity_token
    end
  end
end
