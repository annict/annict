# frozen_string_literal: true

RSpec.configure do |config|
  config.before :suite do
    ImageHelper.module_eval do
      def ann_image_url(*)
        "#{ENV.fetch("ANNICT_URL")}/dummy_image"
      end
      alias_method :v4_ann_image_url, :ann_image_url
    end
  end
end
