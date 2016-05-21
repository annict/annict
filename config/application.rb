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
  action_cable/engine
  sprockets/railtie
).each do |railtie|
  require railtie
end

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(*Rails.groups)

module Annict
  class Application < Rails::Application
    # Settings in config/environments/* take precedence over those specified here.
    # Application configuration should go into files in config/initializers
    # -- all .rb files in that directory are automatically loaded.

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
    config.i18n.available_locales = ["en-US", :ja]

    config.autoload_paths += %W(
      #{config.root}/lib
    )
    config.eager_load_paths += %W(
      #{config.root}/lib
    )

    config.generators do |g|
      g.test_framework :rspec, controller_specs: false, helper_specs: false,
                               routing_specs: false, view_specs: false
      g.factory_girl false
    end

    # `after_rollback`/`after_commit` 内でエラーが発生したときロールバックする
    # Rails 4.2より後のバージョンでこの挙動がデフォルトになる
    config.active_record.raise_in_transactional_callbacks = true

    config.active_job.queue_adapter = :delayed_job

    config.middleware.insert_before(Rack::Runtime, Rack::Rewrite) do
      r301 %r{\A/users/([A-Za-z0-9_]+)\z}, "/@$1"
      # rubocop:disable Metrics/LineLength
      r301 %r{\A/users/([A-Za-z0-9_]+)/(following|followers|wanna_watch|watching|watched|on_hold|stop_watching)\z}, "/@$1/$2"
      # rubocop:enable Metrics/LineLength
    end

    unless Rails.env.test?
      config.cache_store = :redis_store,
        "#{ENV.fetch('REDIS_URL')}/0/annict_cache",
        { expires_in: 90.minutes }
    end

    commandline_options = "-t coffeeify --extension=\".js.coffee\""
    config.browserify_rails.commandline_options = commandline_options
  end
end
