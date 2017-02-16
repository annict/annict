# frozen_string_literal: true

module Annict
  module Amazon
    class Item
      def initialize(asin, country: "jp")
        @asin = asin
        @country = country
        @ecs ||= ::Amazon::Ecs.configure do |options|
          options[:AWS_access_key_id] = ENV.fetch("AWS_PAA_ACCESS_KEY_ID")
          options[:AWS_secret_key] = ENV.fetch("AWS_PAA_SECRET_KEY")
          options[:associate_tag] = ENV.fetch("AMAZON_ASSOCIATE_TAG")
        end
      end

      def detail
        res = ::Amazon::Ecs.item_lookup(@asin,
          country: @country,
          response_group: "Small,Images")
        binding.pry
        image_set_list = res.doc.css("Items Item ImageSets ImageSet")
        image_set_list.map do |image_set|
          image = image_set.css("HiResImage").presence ||
            image_set.css("LargeImage").presence ||
            image_set.css("MediumImage").presence ||
            image_set.css("TinyImage").presence ||
            image_set.css("ThumbnailImage").presence ||
            image_set.css("SmallImage").presence ||
            image_set.css("SwatchImage").presence
          {
            url: image.css("URL").text,
            height: image.css("Height").text.to_i,
            width: image.css("Width").text.to_i
          }
        end
      end
    end
  end
end
