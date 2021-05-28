# frozen_string_literal: true

module V3::Oauth
  class ApplicationController < ActionController::Base
    include Doorkeeper::Helpers::Controller

    include V3::ControllerCommon
    include V3::ViewSelector
    include V6::KeywordSearchable
    include V6::Localizable
    include V6::Loggable
    include V6::PageCategorizable
    include V6::SentryLoadable

    layout "v3/default"

    around_action :set_locale
  end
end
