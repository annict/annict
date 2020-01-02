# frozen_string_literal: true

Rails.application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  # Code is not reloaded between requests.
  config.cache_classes = true

  # Eager load code on boot. This eager loads most of Rails and
  # your application in memory, allowing both threaded web servers
  # and those relying on copy on write to perform better.
  # Rake tasks automatically ignore this option for performance.
  config.eager_load = true

  # Full error reports are disabled and caching is turned on.
  config.consider_all_requests_local       = false
  config.action_controller.perform_caching = true

  # Ensures that a master key has been made available in either ENV["RAILS_MASTER_KEY"]
  # or in config/master.key. This key is used to decrypt credentials (and other encrypted files).
  # config.require_master_key = true

  # Disable serving static files from the `/public` folder by default since
  # Apache or NGINX already handles this.
  # Heroku will set `RAILS_SERVE_STATIC_FILES` when you deploy a Ruby app via
  # the Heroku Ruby Buildpack for Rails 4.2+ apps.
  # https://blog.heroku.com/container_ready_rails_5#serving-files-by-default
  config.public_file_server.enabled = ENV["RAILS_SERVE_STATIC_FILES"].present?

  # Enable serving of images, stylesheets, and JavaScripts from an asset server.
  # `no-image.jpg` はTombo経由で表示するため、`image_path("no-image.jpg")` の返り値に
  # CloudFrontのURLを付加しないようにする
  config.action_controller.asset_host = proc do |source|
    if source =~ %r{/assets/no-image}
      nil
    else
      ENV.fetch("ANNICT_ASSET_URL")
    end
  end

  # Enable Rack::Cache to put a simple HTTP cache in front of your application
  # Add `rack-cache` to your Gemfile before enabling this.
  # For large-scale production use, consider using a caching reverse proxy
  # like nginx, varnish or squid.
  # config.action_dispatch.rack_cache = true

  # Specifies the header that your server uses for sending files.
  # config.action_dispatch.x_sendfile_header = "X-Sendfile" # for Apache
  # config.action_dispatch.x_sendfile_header = 'X-Accel-Redirect' # for NGINX

  # Force all access to the app over SSL, use Strict-Transport-Security,
  # and use secure cookies.
  config.force_ssl = ENV["ANNICT_FORCE_SSL"].present?

  # Use the lowest log level to ensure availability of diagnostic information
  # when problems arise.
  config.log_level = ENV.fetch("ANNICT_LOG_LEVEL", :info)

  # Prepend all log lines with the following tags.
  config.log_tags = [:request_id]

  config.lograge.enabled = true
  config.lograge.formatter = Lograge::Formatters::Json.new
  config.lograge.custom_options = lambda do |event|
    options = event.payload.slice(:request_id, :client_uuid, :user_id)
    options[:params] = event.payload[:params].except("controller", "action")
    options
  end
  config.log_formatter = config.lograge.formatter

  # Use a different cache store in production.
  # config.cache_store = :mem_cache_store

  # Use a real queuing backend for Active Job (and separate queues per environment).
  # config.active_job.queue_adapter     = :resque
  # config.active_job.queue_name_prefix = "rails-5-2_#{Rails.env}"

  config.action_mailer.perform_caching = false

  # Ignore bad email addresses and do not raise email delivery errors.
  # Set this to true and configure the email server for immediate delivery to raise delivery errors.
  # config.action_mailer.raise_delivery_errors = false

  config.action_mailer.asset_host = config.action_controller.asset_host

  # Enable locale fallbacks for I18n (makes lookups for any locale fall back to
  # the I18n.default_locale when a translation cannot be found).
  config.i18n.fallbacks = true

  # Send deprecation notices to registered listeners.
  config.active_support.deprecation = :notify

  config.action_mailer.default_url_options = {
    protocol: "https://",
    host: ENV.fetch("ANNICT_HOST")
  }
  config.action_mailer.delivery_method = :smtp
  config.action_mailer.smtp_settings = {
    address: ENV.fetch("SMTP_HOST"),
    port: ENV.fetch("SMTP_PORT"),
    user_name: ENV.fetch("SMTP_USERNAME"),
    password: ENV.fetch("SMTP_PASSWORD"),
    authentication: :plain
  }

  # Do not dump schema after migrations.
  config.active_record.dump_schema_after_migration = false

  # Inserts middleware to perform automatic connection switching.
  # The `database_selector` hash is used to pass options to the DatabaseSelector
  # middleware. The `delay` is used to determine how long to wait after a write
  # to send a subsequent read to the primary.
  #
  # The `database_resolver` class is used by the middleware to determine which
  # database is appropriate to use based on the time delay.
  #
  # The `database_resolver_context` class is used by the middleware to set
  # timestamps for the last write to the primary. The resolver uses the context
  # class timestamps to determine how long to wait before reading from the
  # replica.
  #
  # By default Rails will store a last write timestamp in the session. The
  # DatabaseSelector middleware is designed as such you can define your own
  # strategy for connection switching and pass that into the middleware through
  # these configuration options.
  # config.active_record.database_selector = { delay: 2.seconds }
  # config.active_record.database_resolver = ActiveRecord::Middleware::DatabaseSelector::Resolver
  # config.active_record.database_resolver_context = ActiveRecord::Middleware::DatabaseSelector::Resolver::Session

  config.middleware.insert_before(Rack::Runtime, Rack::Rewrite) do
    # https://annict.com/sitemap.xml.gz でS3にアップロードされてるサイトマップを取得する
    r301 %r{\A/(sitemaps.*)}, "#{ENV.fetch('ANNICT_SITEMAP_URL')}/$1"
  end

  config.imgix = {
    use_https: true,
    source: ENV.fetch("IMGIX_SOURCE"),
    secure_url_token: ENV.fetch("IMGIX_SECURE_URL_TOKEN")
  }
end
