base_options = {
  hash_secret: ENV["PAPERCLIP_RANDOM_SECRET"],
  styles: { master: ["1000x1000\>", :jpg] },
  convert_options: { master: "-quality 90 -strip" }
}

options = if !Rails.env.development?
  base_options.merge(
    path: ":rails_root/public/#{ENV['PAPERCLIP_PATH']}"
  )
else
  base_options.merge(
    storage: :s3,
    path: ENV["PAPERCLIP_PATH"],
    s3_credentials: {
      # bucket:            ENV["S3_BUCKET_NAME_PRODUCTION"],
      bucket:            ENV["S3_BUCKET_NAME_DEVELOPMENT"],
      access_key_id:     ENV["S3_ACCESS_KEY_ID"],
      secret_access_key: ENV["S3_SECRET_ACCESS_KEY"]
    }
  )
end

Paperclip::Attachment.default_options.update(options)
