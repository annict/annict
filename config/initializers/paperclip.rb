base_options = {
  hash_secret: ENV["ANNICT_PAPERCLIP_RANDOM_SECRET"],
  styles: { master: ["1000x1000\>", :jpg] },
  convert_options: { master: "-quality 90 -strip" }
}

if Rails.env.development? || Rails.env.production?
  options = base_options.merge(
    storage: :s3,
    path: ENV["ANNICT_PAPERCLIP_PATH"],
    s3_credentials: {
      bucket: ENV["S3_BUCKET_NAME"],
      access_key_id: ENV["S3_ACCESS_KEY_ID"],
      secret_access_key: ENV["S3_SECRET_ACCESS_KEY"]
    }
  )
else
  options = base_options.merge(
    path: ":rails_root/spec/test_files/#{ENV['ANNICT_PAPERCLIP_PATH']}"
  )
end

Paperclip::Attachment.default_options.update(options)
