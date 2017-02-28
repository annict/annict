# frozen_string_literal: true

class EmailNotificationMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"
  add_template_helper ImageHelper

  def followed_user(user_id, following_user_id)
    @user = User.find(user_id)
    @following_user = User.find(following_user_id)

    mail(to: @following_user.email, subject: "Hello") do |format|
      format.mjml
    end
  end
end
