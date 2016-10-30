# frozen_string_literal: true

class RegistrationsController < Devise::RegistrationsController
  permits :username, :email, :password, model_name: "User"

  before_action :load_i18n, only: %i(new create)

  def new
    @user = User.new_with_session({}, session)
  end

  def create(user)
    @user = User.new(user).build_relations

    @user.save
    return render(:new) unless @user.valid?

    ga_client.events.create("users", "create")
    bypass_sign_in(@user)

    flash[:info] = t("registrations.create.confirmation_mail_has_sent")
    redirect_to after_sign_in_path_for(@user)
  end

  private

  def load_i18n
    keys = {
      "messages.registrations.new.username_preview": {
        mobile: "messages.registrations.new.username_preview_mobile"
      }
    }
    load_i18n_into_gon keys
  end
end
