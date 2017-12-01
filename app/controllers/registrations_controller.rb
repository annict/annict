# frozen_string_literal: true

class RegistrationsController < Devise::RegistrationsController
  permits :username, :email, :password, model_name: "User"

  def new
    @new_user = User.new_with_session({}, session)
  end

  def create(user)
    @new_user = User.new(user).build_relations
    @new_user.time_zone = cookies["ann_time_zone"].presence || "Asia/Tokyo"
    @new_user.locale = locale

    return render(:new) unless @new_user.valid?

    ActiveRecord::Base.transaction do
      @new_user.save!
      ga_client.user = @new_user
      ga_client.events.create(:users, :create)
      keen_client.publish(
        "create_users",
        user: @new_user,
        via: "web",
        via_oauth: false
      )
    end

    bypass_sign_in(@new_user)

    flash[:notice] = t("messages.registrations.create.confirmation_mail_has_sent")
    redirect_to after_sign_in_path_for(@new_user)
  end
end
