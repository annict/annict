# frozen_string_literal: true

module ControllerCommon
  extend ActiveSupport::Concern

  included do
    helper_method :render_jb, :discord_invite_url, :localable_resources, :browser, :device_pc?

    rescue_from ActionView::MissingTemplate do
      raise ActionController::RoutingError, "Not Found" if Rails.env.production?
      raise
    end

    def browser
      ua = request.headers["User-Agent"]
      @browser ||= Browser.new(ua, accept_language: request.headers["Accept-Language"])
    end

    def device_pc?
      !(browser.device.mobile? || browser.device.tablet?)
    end

    def render_jb(path, assigns)
      ApplicationController.render("#{path}.jb", assigns: assigns)
    end

    def localable_resources(resources)
      if user_signed_in?
        localable_resources_with_own(resources)
      elsif !user_signed_in? && locale_en?
        resources.with_locale(:en)
      elsif !user_signed_in? && locale_ja?
        resources.with_locale(:ja)
      else
        resources
      end
    end

    def localable_resources_with_own(resources)
      collection = resources.with_locale(*current_user.allowed_locales)

      return resources.where(user: current_user).or(collection) if resources.column_names.include?("user_id")

      case resources.first.class.name
      when "UserlandProject"
        UserlandProject.where(id: collection.pluck(:id) + resources.merge(current_user.userland_projects).pluck(:id))
      else
        collection
      end
    end

    private

    def discord_invite_url
      ENV.fetch("ANNICT_DISCORD_INVITE_URL")
    end
  end
end
