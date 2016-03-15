# frozen_string_literal: true

class RegistrationsController < Devise::RegistrationsController
  before_action :set_oauth, only: [:new, :create]

  def new
    username = @oauth[:info][:nickname].presence || ""
    email = @oauth[:info][:email].presence || ""

    # Facebookからのユーザ登録のとき `username` に「.」が
    # 含まれている可能性があるので除去する
    username = username.tr(".", "_")

    @user = User.new(username: username, email: email)

    render layout: "v1/application"
  end

  def create
    @user = User.new(user_params).build_relations(@oauth)

    if @user.save
      sign_in(@user, bypass: true)

      flash[:info] = t("registrations.create.confirmation_mail_has_sent")
      redirect_to after_sign_in_path_for(@user)
    else
      render :new, layout: "v1/application"
    end
  end

  private

  def user_params
    params.require(:user).permit(:email, :username)
  end

  def set_oauth
    @oauth = session["devise.oauth_data"]
  end
end
