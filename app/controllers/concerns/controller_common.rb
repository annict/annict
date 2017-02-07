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
        I18n.locale = preferred_languages.include?("ja") ? :ja : :en
      end
    end

    def locale_subdomain?
      I18n.available_locales.include?(request.subdomain&.to_sym)
    end
  end
end
