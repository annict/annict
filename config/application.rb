require File.expand_path('../boot', __FILE__)

# Pick the frameworks you want:
require "active_record/railtie"
require "action_controller/railtie"
require "action_mailer/railtie"
require "sprockets/railtie"
# require "rails/test_unit/railtie"

# Require the gems listed in Gemfile, including any gems
# you've limited to :test, :development, or :production.
Bundler.require(:default, Rails.env)

module Annict
  class Application < Rails::Application
    # Settings in config/environments/* take precedence over those specified here.
    # Application configuration should go into files in config/initializers
    # -- all .rb files in that directory are automatically loaded.

    # Set Time.zone default to the specified zone and make Active Record auto-convert to this zone.
    # Run "rake -D time" for a list of tasks for finding time zone names. Default is UTC.
    # config.time_zone = 'Asia/Tokyo'
    # config.active_record.default_timezone = :local

    config.i18n.enforce_available_locales = false

    # The default locale is :en and all translations from config/locales/*.rb,yml are auto loaded.
    # config.i18n.load_path += Dir[Rails.root.join('my', 'locales', '*.{rb,yml}').to_s]
    config.i18n.default_locale = :ja
    config.i18n.available_locales = ['en-US', :ja]

    config.autoload_paths += %W(
      #{config.root}/lib
      #{config.root}/app/forms
      #{config.root}/app/decorators/concerns
    )

    ['fonts'].each do |dir_name|
      config.assets.paths << "#{Rails.root}/app/assets/#{dir_name}"
    end

    config.generators do |g|
      g.test_framework :rspec, controller_specs: false, helper_specs: false,
                               routing_specs: false, view_specs: false
      g.factory_girl false
    end

    # `after_rollback`/`after_commit` 内でエラーが発生したときロールバックする
    # Rails 4.2より後のバージョンでこの挙動がデフォルトになる
    config.active_record.raise_in_transactional_callbacks = true

    config.active_job.queue_adapter = :delayed_job
  end
end
