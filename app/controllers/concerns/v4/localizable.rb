# frozen_string_literal: true

module V4
  module Localizable
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

    def locale_ja?
      locale.to_s == "ja"
    end

    def locale_en?
      locale.to_s == "en"
    end

    def local_time_zone
      current_user&.time_zone.presence || cookies["ann_time_zone"]
    end
  end
end
