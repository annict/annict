Annict::Application.configure do
  # Settings specified here will take precedence over those in config/application.rb.

  # Code is not reloaded between requests.
  config.cache_classes = true

  # Eager load code on boot. This eager loads most of Rails and
  # your application in memory, allowing both thread web servers
  # and those relying on copy on write to perform better.
  # Rake tasks automatically ignore this option for performance.
  config.eager_load = true

  # Full error reports are disabled and caching is turned on.
  config.consider_all_requests_local       = false
  config.action_controller.perform_caching = true

  # Enable Rack::Cache to put a simple HTTP cache in front of your application
  # Add `rack-cache` to your Gemfile before enabling this.
  # For large-scale production use, consider using a caching reverse proxy like nginx, varnish or squid.
  # config.action_dispatch.rack_cache = true

  # Disable Rails's static asset server (Apache or nginx will already do this).
  config.serve_static_files = false

  # Compress JavaScripts and CSS.
  config.assets.js_compressor = :uglifier
  # config.assets.css_compressor = :sass

  # Do not fallback to assets pipeline if a precompiled asset is missed.
  config.assets.compile = false

  # Generate digests for assets URLs.
  config.assets.digest = true

  # Version of your assets, change this if you want to expire all your assets.
  config.assets.version = '1.0'

  # Specifies the header that your server uses for sending files.
  # config.action_dispatch.x_sendfile_header = "X-Sendfile" # for apache
  # config.action_dispatch.x_sendfile_header = 'X-Accel-Redirect' # for nginx

  # Force all access to the app over SSL, use Strict-Transport-Security, and use secure cookies.
  config.force_ssl = true

  # Set to :debug to see everything in the log.
  config.log_level = :info

  # Prepend all log lines with the following tags.
  # config.log_tags = [ :subdomain, :uuid ]

  # Use a different logger for distributed setups.
  # config.logger = ActiveSupport::TaggedLogging.new(SyslogLogger.new)

  # Use a different cache store in production.
  # config.cache_store = :mem_cache_store

  # Enable serving of images, stylesheets, and JavaScripts from an asset server.
  # `no-image.jpg` はTombo経由で表示するため、`image_path("no-image.jpg")` の返り値に
  # CloudFrontのURLを付加しないようにする
  config.action_controller.asset_host = Proc.new do |source, request|
    if %r(/assets/no-image) === source
      nil
    else
      "https://d3a8d1smk6xli.cloudfront.net"
    end
  end

  # Precompile additional assets.
  # application.js, application.css, and all non-JS/CSS in app/assets folder are already added.
  config.assets.precompile += %w(
    application_mobile.scss
    application_common.js.coffee
    db.scss
  )

  # Ignore bad email addresses and do not raise email delivery errors.
  # Set this to true and configure the email server for immediate delivery to raise delivery errors.
  # config.action_mailer.raise_delivery_errors = false

  # Enable locale fallbacks for I18n (makes lookups for any locale fall back to
  # the I18n.default_locale when a translation can not be found).
  config.i18n.fallbacks = true

  # Send deprecation notices to registered listeners.
  config.active_support.deprecation = :notify

  # Disable automatic flushing of the log to improve performance.
  # config.autoflush_log = false

  # Use default logging formatter so that PID and timestamp are not suppressed.
  config.log_formatter = ::Logger::Formatter.new

  # Prevents to break some of JavaScript (ex: Angular.js) code due to renaming
  # https://shellycloud.com/blog/2013/10/how-to-integrate-angularjs-with-rails-4
  config.assets.js_compressor = Uglifier.new(mangle: false)

  config.action_mailer.default_url_options = { protocol: 'https://', host: ENV['ANNICT_HOST'] }
  config.action_mailer.delivery_method = :smtp
  config.action_mailer.smtp_settings = {
      address:        ENV['SMTP_HOST'],
      port:           ENV['SMTP_PORT'],
      user_name:      ENV['SMTP_USERNAME'],
      password:       ENV['SMTP_PASSWORD'],
      authentication: :plain
  }

  config.middleware.insert_before(Rack::Runtime, Rack::Rewrite) do
    # http://www.annict.com へのリクエストを https://annict.com にリダイレクトする
    r301 %r{.*}, "https://annict.com$&", if: Proc.new { |rack_env|
      rack_env["SERVER_NAME"] == "www.annict.com"
    }
    # https://annict.com/sitemap.xml.gz でS3にアップロードされてるサイトマップを取得する
    r301 %r{/sitemap.xml.gz}, ENV.fetch("ANNICT_SITEMAP_URL")
  end

  config.imgix = {
    secure: true,
    source: ENV.fetch("IMGIX_SOURCE")
  }
end
