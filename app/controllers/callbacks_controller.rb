class CallbacksController < Devise::OmniauthCallbacksController
  before_filter :authorize, only: [:facebook, :twitter]

  def twitter
  end

  def facebook
  end


  private

  def authorize
    auth = request.env['omniauth.auth']
    provider = Provider.find_by(name: auth[:provider], uid: auth[:uid])

    if provider.present?
      provider.token            = auth[:credentials][:token]
      provider.token_expires_at = auth[:credentials][:expires_at] if 'facebook' == auth[:provider]
      provider.token_secret     = auth[:credentials][:secret]     if 'twitter' == auth[:provider]
      provider.save

      sign_in_and_redirect(provider.user, event: :authentication)
    else
      session['devise.oauth_data'] = auth
      redirect_to new_user_registration_path
    end
  end
end