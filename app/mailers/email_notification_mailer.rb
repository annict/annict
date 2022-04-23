# frozen_string_literal: true

class EmailNotificationMailer < ApplicationMailer
  helper :local
  helper :image

  def followed_user(followed_user_id, user_id)
    @followed_user = User.only_kept.find(followed_user_id)
    @unsubscription_key = @followed_user.email_notification.unsubscription_key
    @user = User.only_kept.find(user_id)

    I18n.with_locale(@followed_user.locale) do
      subject = default_i18n_subject(
        name: @user.profile.name,
        username: @user.username
      )
      mail(to: @followed_user.email, subject: subject, &:mjml)
    end
  end

  def liked_episode_record(liked_user_id, user_id, episode_record_id)
    @liked_user = User.only_kept.find(liked_user_id)
    @unsubscription_key = @liked_user.email_notification.unsubscription_key
    @user = User.only_kept.find(user_id)
    @episode_record = @liked_user.episode_records.only_kept.find(episode_record_id)
    @work = @episode_record.work
    @episode = @episode_record.episode

    I18n.with_locale(@liked_user.locale) do
      subject = default_i18n_subject(
        name: @user.profile.name,
        username: @user.username,
        title: @episode.decorate.number_with_work_title
      )
      mail(to: @liked_user.email, subject: subject, &:mjml)
    end
  end

  def favorite_works_added(user_id, work_ids)
    @user = User.only_kept.find(user_id)
    @unsubscription_key = @user.email_notification.unsubscription_key
    @works = Work.only_kept.where(id: work_ids)

    I18n.with_locale(@user.locale) do
      subject = default_i18n_subject(n: @works.size)
      mail(to: @user.email, subject: subject, &:mjml)
    end
  end

  def related_works_added(user_id, work_ids)
    @user = User.only_kept.find(user_id)
    @unsubscription_key = @user.email_notification.unsubscription_key
    @works = Work.only_kept.where(id: work_ids)

    I18n.with_locale(@user.locale) do
      subject = default_i18n_subject(n: @works.size)
      mail(to: @user.email, subject: subject, &:mjml)
    end
  end
end
