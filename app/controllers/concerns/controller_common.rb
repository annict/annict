# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
    helper_method :render_jb

    if ENV.fetch("ANNICT_BASIC_AUTH") == "on"
      name = ENV.fetch("ANNICT_BASIC_AUTH_NAME")
      password = ENV.fetch("ANNICT_BASIC_AUTH_PASSWORD")
      http_basic_authenticate_with name: name, password: password
    end

    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
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
      if user_signed_in? && locale_subdomain?
        locale = request.subdomain
        return redirect_to_root_domain if current_user.locale == locale
        current_user.update_column(:locale, locale)
        redirect_to_root_domain
      elsif user_signed_in? && !locale_subdomain?
        I18n.locale = current_user.locale
      elsif !user_signed_in? && locale_subdomain?
        I18n.locale = request.subdomain
      else
        preferred_languages = http_accept_language.user_preferred_languages
        # Chrome returns "ja", but Safari would return "ja-JP", not "ja".
        I18n.locale = if preferred_languages.any? { |lang| lang.match?(/ja/) }
          :ja
        else
          :en
        end
      end
    end

    def locale_subdomain?
      I18n.available_locales.include?(request.subdomain&.to_sym)
    end
  end
end
