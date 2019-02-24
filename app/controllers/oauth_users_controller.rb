# frozen_string_literal: true

class OauthUsersController < Devise::RegistrationsController
  permits :username, :email, :terms_and_privacy_policy_agreement, model_name: "User"

  before_action :set_oauth, only: %i(new create)

  def new
    username = @oauth[:info][:nickname].presence || ""
    email = @oauth[:info][:email].presence || ""

    # Facebookからのユーザ登録のとき `username` に「.」が
    # 含まれている可能性があるので除去する
    username = username.tr(".", "_")

    @user = User.new(username: username, email: email)
  end

  def create(user)
    @user = User.new(user).build_relations(@oauth)
    @user.time_zone = cookies["ann_time_zone"].presence || "Asia/Tokyo"
    @user.locale = locale

    return render(:new) unless @user.valid?

    @user.save
    ga_client.user = @user
    ga_client.events.create(:users, :create, el: "via_oauth")
    keen_client.publish(:user_create, via: "web", via_oauth: true)

    bypass_sign_in(@user)

    flash[:notice] = t("messages.registrations.create.confirmation_mail_has_sent")
    redirect_to after_sign_in_path_for(@user)
  end

  private

  def set_oauth
    @oauth = session["devise.oauth_data"]
    redirect_to new_user_registration_path if @oauth.blank?
  end
end
