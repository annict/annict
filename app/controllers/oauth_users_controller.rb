# frozen_string_literal: true

class OauthUsersController < Devise::RegistrationsController
  before_action :set_oauth, only: %i(new create)

  def new
    username = @oauth["info"]["nickname"].presence || ""
    email = @oauth["info"]["email"].presence || ""

    # Facebookからのユーザ登録のとき `username` に「.」が
    # 含まれている可能性があるので除去する
    username = username.tr(".", "_")

    @user = User.new(username: username, email: email)
  end

  def create
    @user = User.new(user_params).build_relations(@oauth)
    @user.time_zone = cookies["ann_time_zone"].presence || "Asia/Tokyo"
    @user.locale = locale

    return render(:new) unless @user.valid?

    @user.setting.privacy_policy_agreed = true
    @user.save
    ga_client.user = @user
    ga_client.events.create(:users, :create, el: "via_oauth")

    flash[:notice] = t("messages.registrations.create.confirmation_mail_has_sent")
    redirect_to root_path
  end

  private

  def set_oauth
    @oauth = session["devise.oauth_data"]
    redirect_to new_user_registration_path if @oauth.blank?
  end

  def user_params
    params.require(:user).permit(:username, :email, :terms_and_privacy_policy_agreement)
  end
end
