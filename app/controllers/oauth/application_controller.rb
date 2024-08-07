# typed: false
# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include PageCategorizable
    include SentryLoadable
    include Loggable
    include Localizable
    include KeywordSearchable

    layout "main_default"

    around_action :set_locale

    def lograge_payload
      {
        user_id: current_user&.id
      }
    end
  end
end
