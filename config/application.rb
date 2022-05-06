# frozen_string_literal: true

require_relative "boot"

require "rails"
require "active_model/railtie"
require "active_job/railtie"
require "active_record/railtie"
require "action_controller/railtie"
require "action_mailer/railtie"
require "action_view/railtie"

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module Annict
  class Application < Rails::Application
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 6.1

    # Configuration for the application, engines, and railties goes here.
    #
    # These settings can be overridden in specific environments using the files
    # in config/environments, which are processed later.
    #
    # config.time_zone = "Central Time (US & Canada)"
    # config.eager_load_paths << Rails.root.join("extras")

    # Don't generate system test files.
    config.generators.system_tests = nil

    # Heroku will set `RAILS_LOG_TO_STDOUT` when you deploy a Ruby app via
    # the Heroku Ruby Buildpack for Rails 4.2+ apps.
    # https://blog.heroku.com/container_ready_rails_5#stdout-logging
    if ENV["RAILS_LOG_TO_STDOUT"].present?
      config.logger = ActiveSupport::Logger.new(STDOUT)
    end

    config.i18n.enforce_available_locales = false

    # The default locale is :en and all translations from
    # config/locales/*.rb,yml are auto loaded.
    # config.i18n.load_path += Dir[Rails.root.join('my', 'locales', '*.{rb,yml}').to_s]
    config.i18n.default_locale = :ja
    config.i18n.available_locales = %i[ja en]

    config.asset_host = ENV.fetch("ANNICT_ASSET_URL")

    config.generators do |g|
      g.test_framework :rspec, controller_specs: false, helper_specs: false,
                               routing_specs: false, view_specs: false
      g.factory_bot false
    end

    config.active_job.queue_adapter = :delayed_job

    config.middleware.insert_before(Rack::Runtime, Rack::Rewrite) do
      # Redirect: www.annict.com, ja.annict.com, jp.annict.com -> annict.com
      r301 /.*/, "https://#{ENV.fetch('ANNICT_HOST')}$&", if: proc { |rack_env|
        rack_env["SERVER_NAME"].in?(["www.#{ENV.fetch('ANNICT_HOST')}", "ja.annict.com", "jp.annict.com"])
      }
      r301 %r{\A/about}, "/"
      r301 %r{\A/activities}, "/"
      r301 %r{\A/programs}, "/track"
      r301 %r{\A/users/([A-Za-z0-9_]+)\z}, "/@$1"
      r301 %r{\A/users/([A-Za-z0-9_]+)/(following|followers|wanna_watch|watching|watched|on_hold|stop_watching)\z}, "/@$1/$2"
      r301 %r{\A/@([A-Za-z0-9_]+)/reviews\z}, "/@$1/records"
      r301 %r{\A/episodes/[0-9]+/items}, "/"
      r301 %r{\A/faqs}, "/faq"
      r301 %r{\A/menu}, "/"
      r301 %r{\A/works\z}, "/works/#{ENV.fetch("ANNICT_CURRENT_SEASON")}"
      r301 %r{\A/works/[0-9]+/items}, "/"

      maintenance_file = File.join(Rails.root, "public", "maintenance.html")
      send_file(/(.*)$(?<!maintenance|favicons)/, maintenance_file, if: proc { |rack_env|
        ip_address = rack_env["HTTP_CF_CONNECTING_IP"]

        File.exist?(maintenance_file) &&
          ENV["ANNICT_MAINTENANCE_MODE"] == "on" &&
          ip_address != ENV["ANNICT_ADMIN_IP"]
      })
    end

    config.middleware.insert_before(0, Rack::Cors) do
      ALLOWED_METHODS = %i[get post patch delete options].freeze
      EXPOSED_HEADERS = %w[ETag].freeze
      allow do
        origins "*"
        resource "*", headers: :any, methods: ALLOWED_METHODS, expose: EXPOSED_HEADERS
      end
    end

    # Gzip all the things
    # https://schneems.com/2017/11/08/80-smaller-rails-footprint-with-rack-deflate/
    config.middleware.insert_after ActionDispatch::Static, Rack::Deflater

    Sentry.init do |config|
      config.dsn = ENV.fetch("SENTRY_DSN")
      config.breadcrumbs_logger = %i[active_support_logger http_logger]

      # Set tracesSampleRate to 1.0 to capture 100%
      # of transactions for performance monitoring.
      # We recommend adjusting this value in production
      config.traces_sample_rate = 0.5

      filter = ActiveSupport::ParameterFilter.new(Rails.application.config.filter_parameters)
      config.before_send = lambda do |event, hint|
        # Use Rails' parameter filter to sanitize the event
        filter.filter(event.to_hash)
      end
    end

    ActiveRecord::SessionStore::Session.serializer = :null
  end
end
