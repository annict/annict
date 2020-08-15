# frozen_string_literal: true

class EmailConfirmationMailer < ActionMailer::Base
  include Localable

  default from: "Annict <no-reply@annict.com>"

  def sign_up_confirmation(email_confirmation_id, locale)
    email_confirmation = EmailConfirmation.find(email_confirmation_id)

    @email = email_confirmation.email
    @url = new_registration_url(token: email_confirmation.token, host: local_url(locale: locale))

    I18n.with_locale(locale) do
      subject = default_i18n_subject
      mail(to: @email, subject: subject)
    end
  end

  def sign_in_confirmation(email_confirmation_id, locale)
    email_confirmation = EmailConfirmation.find(email_confirmation_id)

    @email = email_confirmation.email
    @url = sign_in_callback_url(token: email_confirmation.token, host: local_url(locale: locale))

    I18n.with_locale(locale) do
      subject = default_i18n_subject
      mail(to: @email, subject: subject)
    end
  end
end
