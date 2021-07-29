# frozen_string_literal: true

Rails.application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  # Enable/disable caching. By default caching is disabled.
  is_cache_enabled = Rails.root.join("tmp/caching-dev.txt").exist?

  # In the development environment your application's code is reloaded on
  # every request. This slows down response time but is perfect for development
  # since you don't have to restart the web server when you make code changes.
  config.cache_classes = is_cache_enabled

  # Do not eager load code on boot.
  config.eager_load = false

  # Show full error reports.
  config.consider_all_requests_local = true

  config.action_controller.asset_host = ENV.fetch("ANNICT_ASSET_URL")
  config.action_controller.perform_caching = is_cache_enabled
  config.action_controller.enable_fragment_cache_logging = is_cache_enabled

  if is_cache_enabled
    config.cache_store = :redis_cache_store, {
      url: ENV.fetch("REDIS_URL"),
      expires_in: 1.hour.to_i
    }
    config.graphql_fragment_cache.store = :redis_cache_store, {
      url: ENV.fetch("REDIS_URL"),
      expires_in: 1.hour.to_i
    }
    config.public_file_server.headers = {
      "Cache-Control" => "public, max-age=#{2.days.to_i}"
    }
  else
    config.cache_store = :null_store
  end

  # Don't care if the mailer can't send.
  config.action_mailer.raise_delivery_errors = false
  config.action_mailer.perform_caching = false
  config.action_mailer.default_url_options = {host: ENV.fetch("ANNICT_HOST")}
  config.action_mailer.delivery_method = :smtp
  config.action_mailer.smtp_settings = {
    user_name: ENV.fetch("MAILTRAP_USERNAME"),
    password: ENV.fetch("MAILTRAP_PASSWORD"),
    address: "smtp.mailtrap.io",
    domain: "smtp.mailtrap.io",
    port: "2525",
    authentication: :cram_md5
  }

  # Print deprecation notices to the Rails logger.
  config.active_support.deprecation = :log

  # Raise exceptions for disallowed deprecations.
  config.active_support.disallowed_deprecation = :raise

  # Tell Active Support which deprecation messages to disallow.
  config.active_support.disallowed_deprecation_warnings = []

  # Raise an error on page load if there are pending migrations.
  config.active_record.migration_error = :page_load

  # Highlight code that triggered database queries in logs.
  config.active_record.verbose_query_logs = true

  # Raises error for missing translations
  # config.action_view.raise_on_missing_translations = true

  # Annotate rendered view with file names.
  # config.action_view.annotate_rendered_view_with_filenames = true

  # Use an evented file watcher to asynchronously detect changes in source code,
  # routes, locales, etc. This feature depends on the listen gem.
  config.file_watcher = ActiveSupport::EventedFileUpdateChecker

  config.after_initialize do
    Bullet.enable = true
    Bullet.bullet_logger = true
    Bullet.console = true
    Bullet.rails_logger = true
  end

  # Vagrant環境でもBetter Errorsが使いたい
  # https://github.com/charliesome/better_errors#security
  # Wating to be fixed: https://github.com/charliesome/better_errors/issues/341
  # BetterErrors::Middleware.allow_ip! "192.168.33.1"

  # https://github.com/ruckus/active-record-query-trace
  ActiveRecordQueryTrace.enabled = true

  config.imgix = {
    use_https: true,
    source: ENV.fetch("IMGIX_SOURCE"),
    secure_url_token: ENV.fetch("IMGIX_SECURE_URL_TOKEN")
  }

  config.hosts += [
    ".ngrok.io",
    ENV.fetch("ANNICT_API_DOMAIN"),
    ENV.fetch("ANNICT_DOMAIN"),
    ENV.fetch("ANNICT_EN_DOMAIN"),
    ENV.fetch("ANNICT_JP_DOMAIN")
  ]
end
