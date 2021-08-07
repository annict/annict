# frozen_string_literal: true

module Annict
  module VodImporter
    class AmazonVideo < Annict::VodImporter::Base
      include HTTParty

      base_uri "https://www.amazon.co.jp"

      def self.import
        new.import
      end

      def import
        puts "--- AmazonVideo"
        attrs = []
        channel = Channel.find(Channel::AMAZON_VIDEO_ID)

        1.step do |page|
          query = [
            "rh=n%3A2351649051%2Cn%3A%212351650051%2Cn%3A2478407051%2Cp_n_entity_type%3A4174099051%2Cp_n_ways_to_watch%3A3746328051",
            "sort=date-desc-rank",
            "page=#{page}"
          ].join("&")
          path = "/s/ref=sr_pg_2?#{query}"
          puts "Accessing to: #{path}"
          headers = {"User-Agent": "Mozilla/5.0"}
          response = self.class.get(path, headers: headers)
          result = Nokogiri::HTML(response)
          item_list = result.css(".s-result-item")
          break if item_list.blank?

          attrs << item_list.map { |item|
            name = item.css(".s-access-title").text
            name = name.sub("【TBSオンデマンド】", "")
            name = name.sub("（フジテレビオンデマンド）", "")
            {
              code: item.attr("data-asin"),
              name: name.strip
            }
          }

          vod_title_ids = create_vod_title!(channel, attrs.flatten)
          break if vod_title_ids.all?(&:nil?)

          attrs = []
          sleep rand(1..5)
        end
      end
    end
  end
end
