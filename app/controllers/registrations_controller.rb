class RegistrationsController < Devise::RegistrationsController
  before_filter :set_oauth, only: [:new, :create]

  def new
    username = @oauth[:info][:nickname].presence || ''
    email = @oauth[:info][:email].presence || ''

    @user = User.new(username: username, email: email)
    @user.trim_username!
  end

  def create
    @user = User.new(user_params).build_relations(@oauth)

    if @user.save
      flash[:info] = t('registrations.create.confirmation_mail_has_sent')
      respond_with @user, location: root_path
    else
      render 'new'
    end
  end


  private

  def user_params
    params.require(:user).permit(:email, :username, :terms)
  end

  def set_oauth
    @oauth = session['devise.oauth_data']
  end
end