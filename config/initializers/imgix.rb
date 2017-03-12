# frozen_string_literal: true

if Rails.env.development?
  module Imgix
    module Rails
      module ViewHelper
        # https://github.com/imgix/imgix-rails が提供する `ix_image_url` メソッドを上書きする
        # 開発環境ではDmmyixを利用して画像のリサイズを行う
        def ix_image_url(source, options = {})
          "https://placeholdit.imgix.net/~text?txtsize=33&txt=350%C3%97150&w=350&h=150"
          # "/dmmyix#{source}?#{options.to_query}"
        end
      end
    end
  end
end
