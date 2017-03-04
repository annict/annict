# frozen_string_literal: true

class EmailNotificationMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"
  add_template_helper ImageHelper

  def followed_user(user_id, following_user_id)
    @user = User.find(user_id)
    @following_user = User.find(following_user_id)

    I18n.locale = @following_user.locale

    subject = default_i18n_subject(
      name: @user.profile.name,
      username: @user.username
    )
    mail(to: @following_user.email, subject: subject, &:mjml)
  end
end
