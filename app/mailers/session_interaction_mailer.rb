# frozen_string_literal: true

class SessionInteractionMailer < ActionMailer::Base
  include Localable

  default from: "Annict <no-reply@annict.com>"

  def sign_up_interaction(session_interaction, locale)
    @email = session_interaction.email
    @url = new_registration_url(token: session_interaction.token, host: local_url(locale: locale))

    I18n.with_locale(locale) do
      subject = default_i18n_subject
      mail(to: @email, subject: subject)
    end
  end

  def sign_in_interaction(session_interaction, locale)
    @email = session_interaction.email
    @url = sign_in_callback_url(token: session_interaction.token, host: local_url(locale: locale))

    I18n.with_locale(locale) do
      subject = default_i18n_subject
      mail(to: @email, subject: subject)
    end
  end
end
