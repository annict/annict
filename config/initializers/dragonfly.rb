require 'dragonfly'
require 'dragonfly/s3_data_store'

# Configure
Dragonfly.app.configure do
  plugin :imagemagick

  url_format '/media/:job/:name'
  url_host   ENV['DRAGONFLY_URL_HOST']

  common_s3_settings = {
    access_key_id:     ENV['S3_ACCESS_KEY_ID'],
    secret_access_key: ENV['S3_SECRET_ACCESS_KEY'],
    region:            ENV['S3_REGION']
  }

  if Rails.env.production?
    datastore :s3, { bucket_name: ENV['S3_BUCKET_NAME_PRODUCTION'] }.merge(common_s3_settings)
  elsif Rails.env.development? && ENV['S3_BUCKET_NAME_DEVELOPMENT']
    datastore :s3, { bucket_name: ENV['S3_BUCKET_NAME_DEVELOPMENT'] }.merge(common_s3_settings)
  else
    datastore :file,
      root_path: 'public/dragonfly'
  end
end

# Logger
Dragonfly.logger = Rails.logger

# Mount as middleware
Rails.application.middleware.use Dragonfly::Middleware

# Add model functionality
if defined?(ActiveRecord::Base)
  ActiveRecord::Base.extend Dragonfly::Model
  ActiveRecord::Base.extend Dragonfly::Model::Validations
end