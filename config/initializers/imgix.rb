if Rails.env.development?
  module Imgix
    module Rails
      module ViewHelper
        # https://github.com/imgix/imgix-rails が提供する `ix_image_url` メソッドを上書きする
        # 開発環境ではTomboを利用して画像のリサイズを行う
        def ix_image_url(source, options = {})
          tombo_url = ENV.fetch("ANNICT_TOMBO_URL")
          size = "w:#{options[:w]},h:#{options[:h]}"
          "#{tombo_url}/#{size}/#{source}"
        end
      end
    end
  end
end
