# frozen_string_literal: true

class ApplicationController < ActionController::Base
  include PageCategorizable
  include SentryLoadable
  include Loggable
  include Localizable
  include KeywordSearchable

  layout "default"

  helper_method :client_uuid, :local_url_with_path, :locale_en?, :locale_ja?, :local_url, :page_category

  before_action :set_sentry_context
  before_action :set_search_params
  around_action :set_locale

  private

  def redirect_if_signed_in
    if user_signed_in?
      redirect_to root_path
    end
  end

  # Override `Devise::Controllers::Helpers#signed_in_root_path`
  def signed_in_root_path(_resource_or_scope)
    root_path
  end

  # Override `Devise::Controllers::Helpers#after_sign_out_path_for`
  def after_sign_out_path_for(_resource_or_scope)
    root_path
  end
end
