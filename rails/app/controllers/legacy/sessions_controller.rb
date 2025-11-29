# typed: false
# frozen_string_literal: true

module Legacy
  class SessionsController < Devise::SessionsController
    layout "main_simple"

    def new
      store_location_for(:user, params[:back]) if params[:back].present?
      super
    end

    def create
      super do |user|
        if !user.confirmed? && user.registered_after_email_confirmation_required?
          sign_out
          return redirect_to root_path, alert: t("devise.failure.user.unconfirmed")
        end
      end
    end
  end
end
