# frozen_string_literal: true

class OauthUsersController < Devise::RegistrationsController
  permits :username, :email, model_name: "User"

  before_action :load_i18n, only: %i(new create)
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

    @user.save
    return render(:new) unless @user.valid?

    ga_client.events.create("users", "create")
    bypass_sign_in(@user)

    flash[:info] = t("registrations.create.confirmation_mail_has_sent")
    redirect_to after_sign_in_path_for(@user)
  end

  private

  def set_oauth
    @oauth = session["devise.oauth_data"]
  end

  def load_i18n
    keys = {
      "messages.registrations.new.username_preview": {
        mobile: "messages.registrations.new.username_preview_mobile"
      }
    }
    load_i18n_into_gon keys
  end
end
