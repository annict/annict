# frozen_string_literal: true

class CallbacksController < Devise::OmniauthCallbacksController
  before_action :authorize, only: [:facebook, :twitter]

  def twitter
  end

  def facebook
  end

  private

  def authorize
    auth = request.env["omniauth.auth"]
    provider = Provider.find_by(name: auth[:provider], uid: auth[:uid])

    if provider.present?
      provider.attributes = provider_attributes(auth)
      provider.save
      return sign_in_and_redirect(provider.user, event: :authentication)
    end

    if user_signed_in?
      current_user.providers.create(provider_attributes(auth))
      omni_params = request.env["omniauth.params"]
      redirect_path = omni_params["back"].presence || root_path
      bypass_sign_in(current_user)
      redirect_to redirect_path, notice: "連携しました"
    else
      session["devise.oauth_data"] = auth
      redirect_to new_oauth_user_path
    end
  end

  def provider_attributes(auth)
    credentials = auth[:credentials]

    expires_at = (auth[:provider] == "facebook") ? credentials[:expires_at] : nil
    token_secret = (auth[:provider] == "twitter") ? credentials[:secret] : nil

    {
      name: auth[:provider],
      uid: auth[:uid],
      token: credentials[:token],
      token_expires_at: expires_at,
      token_secret: token_secret,
    }
  end

  def after_omniauth_failure_path_for(scope)
    root_path(scope)
  end
end
