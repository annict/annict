# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
    helper_method :render_jb, :locale_ja?, :locale_en?

    rescue_from ActionView::MissingTemplate do
      raise ActionController::RoutingError, "Not Found" if Rails.env.production?
      raise
    end

    if ENV.fetch("ANNICT_BASIC_AUTH") == "on"
      name = ENV.fetch("ANNICT_BASIC_AUTH_NAME")
      password = ENV.fetch("ANNICT_BASIC_AUTH_PASSWORD")
      http_basic_authenticate_with name: name, password: password
    end

    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
    end

    def locale_ja?
      locale == :ja
    end

    def locale_en?
      locale == :en
    end

    private

    def set_search_params
      @search = SearchService.new(params[:q])
    end

    def load_new_user
      return if user_signed_in?
      @new_user = User.new_with_session({}, session)
    end

    def redirect_to_root_domain(options = {})
      url = ENV.fetch("ANNICT_URL")
      redirect_to "#{url}#{request.path}", options
    end

    def redirect_to_locale_domain(options = {})
      return if request.domain == ENV.fetch("ANNICT_DOMAIN") && I18n.locale.to_s == "en"

      url = case I18n.locale.to_s
      when "ja"
        ENV.fetch("ANNICT_JP_URL")
      else
        ENV.fetch("ANNICT_URL")
      end

      redirect_to "#{url}#{request.path}", options
    end

    def redirect_if_unexpected_subdomain
      return unless %w(api).include?(request.subdomain)

      white_list = [
        "/sign_in",
        "/users/auth/facebook/callback",
        "/users/auth/twitter/callback"
      ]
      return if request.path.in?(white_list)

      redirect_to_root_domain(status: 301)
    end

    def switch_languages
      case request.domain
      when ENV.fetch("ANNICT_DOMAIN")
        if user_signed_in? && current_user.locale.en?
          I18n.locale = :en
          return
        end

        I18n.locale = if params[:locale].in?(User.locale.values)
          params[:locale]
        else
          user_signed_in? ? current_user.locale : preferred_locale
        end

        if user_signed_in? && current_user.locale != I18n.locale.to_s
          current_user.update_column(:locale, I18n.locale)
        end

        redirect_to_locale_domain
      when ENV.fetch("ANNICT_JP_DOMAIN")
        I18n.locale = :ja

        if user_signed_in? && !current_user.locale.ja?
          current_user.update_column(:locale, I18n.locale)
        end
      end
    end

    def preferred_locale
      preferred_languages = http_accept_language.user_preferred_languages
      # Chrome returns "ja", but Safari would return "ja-JP", not "ja".
      preferred_languages.any? { |lang| lang.match?(/ja/) } ? :ja : :en
    end
  end
end
