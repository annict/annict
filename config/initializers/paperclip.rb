base_options = {
  hash_secret: ENV["ANNICT_PAPERCLIP_RANDOM_SECRET"],
  styles: { master: ["1000x1000\>", :jpg] },
  convert_options: { master: "-quality 90 -strip" }
}

options = if Rails.env.production?
  base_options.merge(
    storage: :s3,
    path: ENV["ANNICT_PAPERCLIP_PATH"],
    s3_credentials: {
      bucket: ENV["S3_BUCKET_NAME"],
      access_key_id: ENV["S3_ACCESS_KEY_ID"],
      secret_access_key: ENV["S3_SECRET_ACCESS_KEY"]
    },
    s3_region: "ap-northeast-1"
  )
elsif Rails.env.test?
  base_options.merge(
    path: ":rails_root/spec/test_files/#{ENV['ANNICT_PAPERCLIP_PATH']}"
  )
else
  base_options.merge(
    path: ENV.fetch("ANNICT_PAPERCLIP_PATH")
  )
end

Paperclip::Attachment.default_options.update(options)
