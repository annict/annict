# frozen_string_literal: true

if Rails.env.development?
  module Imgix
    module Rails
      module ViewHelper
        # https://github.com/imgix/imgix-rails が提供する `ix_image_url` メソッドを上書きする
        # 開発環境ではTomboを利用して画像のリサイズを行う
        def ix_image_url(source, options = {})
          tombo_url = ENV.fetch("ANNICT_TOMBO_URL")
          params = "w:#{options[:w]},h:#{options[:h]},b:#{options[:blur] / 10}"
          "#{tombo_url}/#{params}/#{source}"
        end
      end
    end
  end
end
