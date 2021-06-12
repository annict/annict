# frozen_string_literal: true

module Localizable
  extend ActiveSupport::Concern

  SKIP_TO_SET_LOCALE_PATHS = %w[
    /users/auth/gumroad/callback
  ].freeze
  private_constant :SKIP_TO_SET_LOCALE_PATHS

  included do
    around_action :switch_time_zone, if: :current_user

    helper_method :local_url_with_path, :locale_en?, :locale_ja?, :local_url
  end

  private

  def set_locale(&action)
    return yield if request.path.in?(SKIP_TO_SET_LOCALE_PATHS)

    case [request.subdomain, request.domain].select(&:present?).join(".")
    when ENV.fetch("ANNICT_DOMAIN")
      I18n.with_locale(:en, &action)
    else
      I18n.with_locale(:ja, &action)
    end
  end

  def set_locale_with_params(&action)
    locale = current_user&.locale.presence || params[:locale].presence || :ja
    I18n.with_locale(locale, &action)
  end

  def locale_ja?
    locale.to_s == "ja"
  end

  def locale_en?
    locale.to_s == "en"
  end

  def local_url(locale: I18n.locale)
    return ENV.fetch("ANNICT_JP_URL") if locale.to_s == "ja"

    ENV.fetch("ANNICT_URL")
  end

  def local_url_with_path(locale: I18n.locale)
    ["#{local_url(locale: locale)}#{request.path}", request.query_string].select(&:present?).join("?")
  end

  def preferred_locale
    preferred_languages = http_accept_language.user_preferred_languages
    # Chrome returns "ja", but Safari would return "ja-JP", not "ja".
    preferred_languages.any? { |lang| lang.match?(/ja/) } ? :ja : :en
  end

  def switch_time_zone(&block)
    Groupdate.time_zone = current_user.time_zone
    Time.use_zone(current_user.time_zone, &block)
  end
end
