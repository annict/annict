# frozen_string_literal: true

module Annict
  module Amazon
    class ItemSearch
      def initialize(res)
        @res = res
      end

      def total_results
        return if @res.blank?
        @res.doc.css("Items TotalResults").text.to_i
      end

      def total_pages
        return if @res.blank?
        @res.doc.css("Items TotalPages").text.to_i
      end

      def item_page
        return if @res.blank?
        @res.doc.css("Items ItemPage").text.to_i
      end

      def items
        return if @res.blank?
        @res.doc.css("Items Item").map { |i| Annict::Amazon::Item.new(i) }
      end
    end
  end
end
