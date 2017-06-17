# frozen_string_literal: true

if Rails.env.development?
  module Imgix
    module Rails
      module UrlHelper
        # https://github.com/imgix/imgix-rails が提供する `ix_image_url` メソッドを上書きする
        # 開発環境ではDmmyixを利用して画像のリサイズを行う
        def ix_image_url(source, options = {})
          if ::Rails.root.join("tmp/caching-dev.txt").exist?
            "http://via.placeholder.com/350x150"
          else
            "/dmmyix#{source}?#{options.to_query}"
          end
        end
      end
    end
  end
end
