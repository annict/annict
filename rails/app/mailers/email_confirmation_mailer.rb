# typed: false
# frozen_string_literal: true

class EmailConfirmationMailer < ApplicationMailer
  include LocalHelper

  def update_email_confirmation(email_confirmation_id, locale)
    email_confirmation = EmailConfirmation.find(email_confirmation_id)
    user = User.find(email_confirmation.user_id)

    @username = user.username
    @url = settings_email_callback_url(token: email_confirmation.token, host: local_url(locale: locale))

    I18n.with_locale(locale) do
      subject = default_i18n_subject
      mail(to: email_confirmation.email, subject: subject)
    end
  end
end
