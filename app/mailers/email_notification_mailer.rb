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

  def liked_record(user_id, record_id)
    @user = User.find(user_id)
    @record = Checkin.find(record_id)
    @liked_user = @record.user
    @work = @record.work
    @episode = @record.episode

    I18n.locale = @liked_user.locale

    subject = default_i18n_subject(
      name: @user.profile.name,
      username: @user.username,
      title: @episode.decorate.number_with_work_title
    )
    mail(to: @liked_user.email, subject: subject, &:mjml)
  end

  def friend_joined(user_id, friend_user_id)
    @user = User.find(user_id)
    @friend_user = User.find(friend_user_id)

    I18n.locale = @user.locale

    subject = default_i18n_subject(
      name: @friend_user.profile.name,
      username: @friend_user.username
    )
    mail(to: @user.email, subject: subject, &:mjml)
  end
end
