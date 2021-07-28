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
      include V4::Localizable

      helper_method :locale_ja?, :locale_en?, :local_url, :page_category

      around_action :set_locale
      skip_before_action :verify_authenticity_token
    end
  end
end
