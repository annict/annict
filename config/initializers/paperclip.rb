base_options = {
  hash_secret: ENV["ANNICT_PAPERCLIP_RANDOM_SECRET"],
  styles: { master: ["1000x1000\>", :jpg] },
  convert_options: { master: "-quality 90 -strip" },
  url: "#{ENV['ANNICT_TOMBO_URL']}/:size/#{ENV['ANNICT_PAPERCLIP_PATH']}"
}

if Rails.env.development?
  options = base_options.merge(
    path: ":rails_root/public/#{ENV['ANNICT_PAPERCLIP_PATH']}"
  )
elsif Rails.env.test?
  options = base_options.merge(
    path: ":rails_root/spec/test_files/#{ENV['ANNICT_PAPERCLIP_PATH']}"
  )
else
  options = base_options.merge(
    storage: :s3,
    path: ENV["ANNICT_PAPERCLIP_PATH"],
    s3_credentials: {
      bucket: ENV["S3_BUCKET_NAME_PRODUCTION"],
      access_key_id: ENV["S3_ACCESS_KEY_ID"],
      secret_access_key: ENV["S3_SECRET_ACCESS_KEY"]
    }
  )
end

Paperclip::Attachment.default_options.update(options)

module PaperclipUrlPatch
  def url(style_name = default_style, options = {})
    if options.has_key?(:size)
      # Paperclipのurlに含まれる:sizeに指定したサイズ値を挿入する
      # http://stackoverflow.com/questions/4041373/paperclip-custom-path-and-url
      Paperclip.interpolates :size do |attachment, style|
        options[:size]
      end
    end

    super
  end
end

class Paperclip::Attachment
  prepend PaperclipUrlPatch
end
