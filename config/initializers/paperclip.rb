# frozen_string_literal: true

base_options = {
  convert_options: { master: "-quality 90 -strip" },
  hash_secret: ENV.fetch("ANNICT_PAPERCLIP_RANDOM_SECRET"),
  path: ENV.fetch("ANNICT_PAPERCLIP_PATH"),
  styles: { master: ["1000x1000\>", :jpg] },
  url: ENV.fetch("ANNICT_PAPERCLIP_URL")
}

options = if Rails.env.production?
  base_options.merge(
    storage: :s3,
    s3_credentials: {
      bucket: ENV.fetch("S3_BUCKET_NAME"),
      access_key_id: ENV.fetch("S3_ACCESS_KEY_ID"),
      secret_access_key: ENV.fetch("S3_SECRET_ACCESS_KEY")
    },
    s3_region: "ap-northeast-1"
  )
elsif Rails.env.test?
  base_options.merge(
    path: ":rails_root/spec/test_files/#{ENV['ANNICT_PAPERCLIP_PATH']}"
  )
else
  base_options
end

Paperclip::Attachment.default_options.update(options)
