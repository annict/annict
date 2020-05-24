# frozen_string_literal: true

require_relative "boot"

%w(
  rails
  active_model/railtie
  active_job/railtie
  active_record/railtie
  action_controller/railtie
  action_mailer/railtie
  action_view/railtie
  view_component/engine
).each do |railtie|
  require railtie
end

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module Annict
  class Application < Rails::Application
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 6.0

    # Heroku will set `RAILS_LOG_TO_STDOUT` when you deploy a Ruby app via
    # the Heroku Ruby Buildpack for Rails 4.2+ apps.
    # https://blog.heroku.com/container_ready_rails_5#stdout-logging
    if ENV["RAILS_LOG_TO_STDOUT"].present?
      config.logger = ActiveSupport::Logger.new(STDOUT)
    end

    # Don't generate system test files.
    config.generators.system_tests = nil

    # Settings in config/environments/* take precedence over those specified here.
    # Application configuration can go into files in config/initializers
    # -- all .rb files in that directory are automatically loaded after loading
    # the framework and any gems in your application.

    # Set Time.zone default to the specified zone and
    # make Active Record auto-convert to this zone.
    # Run "rake -D time" for a list of tasks for finding time zone names. Default is UTC.
    # config.time_zone = 'UTC'
    # config.active_record.default_timezone = :local

    config.i18n.enforce_available_locales = false

    # The default locale is :en and all translations from
    # config/locales/*.rb,yml are auto loaded.
    # config.i18n.load_path += Dir[Rails.root.join('my', 'locales', '*.{rb,yml}').to_s]
    config.i18n.default_locale = :ja
    config.i18n.available_locales = %i(ja en)

    config.generators do |g|
      g.test_framework :rspec, controller_specs: false, helper_specs: false,
                               routing_specs: false, view_specs: false
      g.factory_bot false
    end

    config.active_job.queue_adapter = :delayed_job

    config.active_record.schema_format = :sql

    config.middleware.insert_before(Rack::Runtime, Rack::Rewrite) do
      # Redirect: annict.herokuapp.com -> annict.com
      r301 /.*/, "https://#{ENV.fetch('ANNICT_HOST')}$&", if: proc { |rack_env|
        rack_env["SERVER_NAME"].include?("annict.herokuapp.com")
      }
      # Redirect: www.annict.com -> annict.com
      r301 /.*/, "https://#{ENV.fetch('ANNICT_HOST')}$&", if: proc { |rack_env|
        rack_env["SERVER_NAME"].in?(["www.#{ENV.fetch('ANNICT_HOST')}"])
      }
      # Redirect: www.annict.jp -> annict.jp
      r301 /.*/, "https://#{ENV.fetch('ANNICT_JP_HOST')}$&", if: proc { |rack_env|
        rack_env["SERVER_NAME"].in?(["www.#{ENV.fetch('ANNICT_JP_HOST')}"])
      }
      r301 %r{\A/activities}, "/"
      r301 %r{\A/users/([A-Za-z0-9_]+)\z}, "/@$1"
      r301 %r{\A/users/([A-Za-z0-9_]+)/(following|followers|wanna_watch|watching|watched|on_hold|stop_watching)\z}, "/@$1/$2"
      r301 %r{\A/@([A-Za-z0-9_]+)/reviews\z}, "/@$1/records"
      r301 %r{\A/episodes/[0-9]+/items}, "/"
      r301 %r{\A/works/[0-9]+/items}, "/"

      maintenance_file = File.join(Rails.root, "public", "maintenance.html")
      send_file /(.*)$(?<!maintenance|favicons)/, maintenance_file, if: proc { |rack_env|
        ip_address = rack_env["HTTP_X_FORWARDED_FOR"]&.split(",")&.last&.strip

        File.exist?(maintenance_file) &&
          ENV["ANNICT_MAINTENANCE_MODE"] == "on" &&
          ip_address != ENV["ANNICT_ADMIN_IP"]
      }
    end

    config.middleware.insert_before(0, Rack::Cors) do
      ALLOWED_METHODS = %i(get post patch delete options).freeze
      EXPOSED_HEADERS = %w(ETag).freeze
      allow do
        origins "*"
        resource "*", headers: :any, methods: ALLOWED_METHODS, expose: EXPOSED_HEADERS
      end
    end

    # Gzip all the things
    # https://schneems.com/2017/11/08/80-smaller-rails-footprint-with-rack-deflate/
    config.middleware.insert_after ActionDispatch::Static, Rack::Deflater

    Raven.configure do |config|
      config.dsn = ENV.fetch("SENTRY_DSN")
      config.sanitize_fields = Rails.application.config.filter_parameters.map(&:to_s)
    end

    ActiveRecord::SessionStore::Session.serializer = :null
  end
end
