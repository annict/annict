# frozen_string_literal: true

module RequestLocalizable
  SKIP_TO_SET_LOCALE_PATHS = %w(
    /users/auth/gumroad/callback
  ).freeze
  private_constant :SKIP_TO_SET_LOCALE_PATHS

  private

  def set_locale(&action)
    return if request.path.in?(SKIP_TO_SET_LOCALE_PATHS)

    case [request.subdomain, request.domain].select(&:present?).join(".")
    when ENV.fetch("ANNICT_DOMAIN")
      I18n.with_locale(:en, &action)
    else
      I18n.with_locale(:ja, &action)
    end
  end
end
