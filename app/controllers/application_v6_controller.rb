# frozen_string_literal: true

class ApplicationV6Controller < ActionController::Base
  include V6::PageCategorizable
  include V6::SentryLoadable
  include V6::Loggable
  include V6::Localizable
  include V6::KeywordSearchable

  layout "v6/default"

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
