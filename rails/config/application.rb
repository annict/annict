# typed: false
# frozen_string_literal: true

require_relative "boot"

require "rails"
# Pick the frameworks you want:
require "active_model/railtie"
require "active_job/railtie"
require "active_record/railtie"
# require "active_storage/engine"
require "action_controller/railtie"
require "action_mailer/railtie"
# require "action_mailbox/engine"
# require "action_text/engine"
require "action_view/railtie"
# require "action_cable/engine"
# require "rails/test_unit/railtie"

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module Annict
  class Application < Rails::Application
    # Initialize configuration defaults for originally generated Rails version.
    config.load_defaults 7.1

    # Please, add to the `ignore` list any other `lib` subdirectories that do
    # not contain `.rb` files, or that should not be reloaded or eager loaded.
    # Common ones are `templates`, `generators`, or `middleware`, for example.
    config.autoload_lib(ignore: %w[assets tasks])

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

    config.active_record.schema_format = :sql

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
      # Redirect: api.annict.com/sign_in -> annict.com/sign_in
      r301 %r{\A/sign_in(\?.*)?}, "https://#{ENV.fetch('ANNICT_DOMAIN')}/sign_in$1", if: proc { |rack_env|
        rack_env["SERVER_NAME"] == ENV.fetch("ANNICT_API_DOMAIN", "")
      }
      # Redirect: api.annict.com/oauth/authorize -> annict.com/oauth/authorize
      r301 %r{\A/oauth/authorize(\?.*)?}, "https://#{ENV.fetch('ANNICT_DOMAIN')}/oauth/authorize$1", if: proc { |rack_env|
        rack_env["SERVER_NAME"] == ENV.fetch("ANNICT_API_DOMAIN", "")
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

      maintenance_file = Rails.public_path.join("maintenance.html").to_s
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

    ActiveRecord::SessionStore::Session.serializer = :null
  end
end
