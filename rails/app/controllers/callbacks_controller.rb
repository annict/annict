# typed: false
# frozen_string_literal: true

class CallbacksController < Devise::OmniauthCallbacksController
  def facebook
    auth = request.env["omniauth.auth"]
    omni_params = request.env["omniauth.params"]
    provider = Provider.find_by(name: auth[:provider], uid: auth[:uid])

    if user_signed_in?
      ActiveRecord::Base.transaction do
        current_user.providers.create!(provider_attributes(auth))

        bypass_sign_in(current_user)
      end

      redirect_path = omni_params["back"].presence || root_path
      redirect_to redirect_path, notice: t("messages._common.connected")
    else
      email = auth.dig("info", "email")

      if provider
        user = provider.user

        ActiveRecord::Base.transaction do
          if !user.confirmed? && user.email == email
            user.confirm
          end

          provider.attributes = provider_attributes(auth)
          provider.save!

          sign_in user
        end
      else
        user = User.find_by(email: email)

        unless user
          return redirect_to(root_path, alert: t("messages.callbacks.sign_in_failed"))
        end

        ActiveRecord::Base.transaction do
          unless user.confirmed?
            user.confirm
          end

          sign_in user
        end
      end

      redirect_path = omni_params["back"].presence || after_sign_in_path_for(user)
      redirect_to redirect_path, notice: t("messages.sign_in_callback.show.signed_in")
    end
  end

  private

  def provider_attributes(auth)
    credentials = auth[:credentials]

    {
      name: auth[:provider],
      uid: auth[:uid],
      token: credentials[:token],
      token_expires_at: (auth[:provider] == "facebook" ? credentials[:expires_at] : nil),
      token_secret: nil
    }
  end

  def after_omniauth_failure_path_for(scope)
    root_path(scope)
  end
end
