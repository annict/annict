# frozen_string_literal: true

module Annict
  module VodImporter
    class BandaiChannel < Annict::VodImporter::Base
      include HTTParty

      base_uri "http://www.b-ch.com"

      def self.import
        new.import
      end

      def import
        puts "--- BandaiChannel"
        path = "/ttl/jpchar_list.php"
        puts "Accessing to: #{path}"
        html = self.class.get(path)

        result = Nokogiri::HTML(html)
        links = result.css(".search-list .ttl-list li a")

        attrs = links.map { |link|
          {
            code: link.attr(:href)[/([0-9]+)/],
            name: link.text.strip
          }
        }

        return if attrs.blank?

        channel = Channel.find(Channel::BANDAI_CHANNEL_ID)
        create_vod_title!(channel, attrs)
      end
    end
  end
end
