# frozen_string_literal: true

class EmailNotificationMailer < ActionMailer::Base
  default(
    from: "Annict <no-reply@annict.com>",
    bcc: "admin+email-notification@annict.com"
  )

  add_template_helper ImageHelper
  add_template_helper LocalHelper

  def followed_user(followed_user_id, user_id)
    @followed_user = User.find(followed_user_id)
    @unsubscription_key = @followed_user.email_notification.unsubscription_key
    @user = User.find(user_id)

    I18n.locale = @followed_user.locale

    subject = default_i18n_subject(
      name: @user.profile.name,
      username: @user.username
    )
    mail(to: @followed_user.email, subject: subject, &:mjml)
  end

  def liked_record(liked_user_id, user_id, record_id)
    @liked_user = User.find(liked_user_id)
    @unsubscription_key = @liked_user.email_notification.unsubscription_key
    @user = User.find(user_id)
    @record = @liked_user.records.find(record_id)
    @work = @record.work
    @episode = @record.episode

    I18n.locale = @liked_user.locale

    subject = default_i18n_subject(
      name: @user.profile.name,
      username: @user.username,
      title: @episode.present? ? @episode.decorate.number_with_work_title : @work.decorate.local_title
    )
    mail(to: @liked_user.email, subject: subject, &:mjml)
  end

  def friends_joined(user_id, provider_name, friend_user_ids)
    @user = User.find(user_id)
    @unsubscription_key = @user.email_notification.unsubscription_key
    @provider_name = provider_name
    @friend_users = User.where(id: friend_user_ids)

    I18n.locale = @user.locale

    subject = I18n.t("email_notification_mailer.friends_joined.subject_#{provider_name}")
    mail(to: @user.email, subject: subject, &:mjml)
  end

  def favorite_works_added(user_id, work_id)
    @user = User.find(user_id)
    @unsubscription_key = @user.email_notification.unsubscription_key
    @work = Work.find(work_id)
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

    I18n.locale = @user.locale

    subject = default_i18n_subject(
      work_title: @work.decorate.local_title,
      resource_name: @resources.first.decorate.local_name
    )
    mail(to: @user.email, subject: subject, &:mjml)
  end

  def related_works_added(user_id, work_id)
    @user = User.find(user_id)
    @work = Work.published.find(work_id)
    @related_works = @work.related_works.published.order_by_season
    @unsubscription_key = @user.email_notification.unsubscription_key

    I18n.locale = @user.locale

    subject = default_i18n_subject(
      work_title: @work.decorate.local_title,
      related_work_title: @related_works.first.decorate.local_title
    )
    mail(to: @user.email, subject: subject, &:mjml)
  end
end
