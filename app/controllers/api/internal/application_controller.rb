# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include V6::Localizable
      include V6::Loggable
      include V6::SentryLoadable
      include V6::PageCategorizable

      skip_before_action :verify_authenticity_token
    end
  end
end
