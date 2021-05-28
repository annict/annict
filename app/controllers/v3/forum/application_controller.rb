# frozen_string_literal: true

module V3::Forum
  class ApplicationController < ActionController::Base
    include Pundit

    include V3::Analyzable
    include V3::ControllerCommon
    include V3::FlashMessage
    include V3::Gonable
    include V3::LogrageSetting
    include V3::ViewSelector
    include V6::KeywordSearchable
    include V6::Localizable
    include V6::PageCategorizable
    include V6::SentryLoadable

    layout "v3/default"

    helper_method :gon

    before_action :store_data_into_gon
    before_action :load_new_user
  end
end
