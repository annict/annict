# frozen_string_literal: true

module My
  class ApplicationController < ActionController::Base
    include PageCategorizable
    include V4::RavenContext
    include V4::Loggable
    include V4::Localizable

    layout "default"

    helper_method :client_uuid, :local_url_with_path, :locale_en?, :locale_ja?, :local_url, :page_category

    before_action :set_raven_context
    around_action :set_locale
  end
end
