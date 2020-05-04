# frozen_string_literal: true

class EmailNotificationMailer < ActionMailer::Base
  default(
    from: "Annict <no-reply@annict.com>",
    bcc: "admin+email-notification@annict.com"
  )

  add_template_helper AssetsHelper
  add_template_helper ImageHelper
  add_template_helper LocalHelper

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

  def favorite_works_added(user_id, work_id)
    @user = User.only_kept.find(user_id)
    @unsubscription_key = @user.email_notification.unsubscription_key
    @work = Work.only_kept.find(work_id)
    @characters = @work.
      characters.
      joins(:favorite_characters).
      merge(@user.favorite_characters)
    @people = @work.people.joins(:favorite_people).merge(@user.favorite_people)
    @orgs = @work.
      organizations.
      joins(:favorite_organizations).
      merge(@user.favorite_organizations)
    @resources = @characters | @people | @orgs

    I18n.with_locale(@user.locale) do
      subject = default_i18n_subject(
        work_title: @work.local_title,
        resource_name: @resources.first.local_name
      )
      mail(to: @user.email, subject: subject, &:mjml)
    end
  end

  def related_works_added(user_id, work_id)
    @user = User.only_kept.find(user_id)
    @work = Work.only_kept.find(work_id)
    @related_works = @work.related_works.only_kept.order_by_season
    @unsubscription_key = @user.email_notification.unsubscription_key

    I18n.with_locale(@user.locale) do
      subject = default_i18n_subject(
        work_title: @work.local_title,
        related_work_title: @related_works.first.local_title
      )
      mail(to: @user.email, subject: subject, &:mjml)
    end
  end
end
