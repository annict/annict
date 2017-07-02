# frozen_string_literal: true

module Annict
  module Amazon
    class Items
      def initialize(country: "jp")
        @country = country
        @response_group = "Medium"
      end

      def search(keyword, search_index: "All", item_page: 1)
        res = ::Amazon::Ecs.item_search(keyword,
          country: @country,
          response_group: @response_group,
          search_index: search_index,
          item_page: item_page)
        Annict::Amazon::ItemSearch.new(res)
      end

      def fetch_by_asin(asin)
        res = ::Amazon::Ecs.item_lookup(asin, country: @country, response_group: @response_group)
        Annict::Amazon::Item.new(res.doc.css("Items Item"))
      end
    end
  end
end
