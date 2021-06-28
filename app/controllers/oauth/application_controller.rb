# frozen_string_literal: true

module Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include ControllerCommon
    include ViewSelector
    include KeywordSearchable
    include Localizable
    include Loggable
    include PageCategorizable
    include SentryLoadable

    layout "v3/default"

    around_action :set_locale
  end
end
