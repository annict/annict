# frozen_string_literal: true

module Annict
  module Amazon
    class Item
      extend Enumerize

      enumerize :category, in: %w(
        All
        Music
        DVD
        VideoGames
        PCHardware
        Electronics
        Books
        Hobbies
        Apparel
      )

      def initialize(item)
        @item = item
      end

      def manufacturer
        @item.css("ItemAttributes Manufacturer").text
      end

      def title
        @item.css("ItemAttributes Title").text
      end

      def detail_page_url
        @item.css("DetailPageURL").text
      end

      def asin
        @item.css("ASIN").text
      end

      def ean
        @item.css("ItemAttributes EAN").text
      end

      def amount
        amount = @item.css("ItemAttributes ListPrice Amount").text
        amount.present? ? amount.to_i : nil
      end

      def currency_code
        @item.css("ItemAttributes ListPrice CurrencyCode").text
      end

      def offer_amount
        amount = @item.css("OfferSummary LowestNewPrice Amount").text
        amount.present? ? amount.to_i : nil
      end

      def offer_currency_code
        @item.css("OfferSummary LowestNewPrice CurrencyCode").text
      end

      def release_date
        @item.css("ItemAttributes ReleaseDate").text
      end

      def images
        image_set_list = @item.css("ImageSets ImageSet")
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
