# frozen_string_literal: true

module Annict
  module VodImporter
    class NicoNicoChannel < Annict::VodImporter::Base
      include HTTParty

      base_uri "http://ch.nicovideo.jp"

      def self.import
        new.import
      end

      def import
        puts "--- NicoNicoChannel"
        attrs = []
        channel = Channel.find(Channel::NICONICO_CHANNEL_ID)

        1.step do |i|
          path = "/portal/anime/list?page=#{i}"

          puts "Accessing to: #{path}"
          html = self.class.get(path)

          result = Nokogiri::HTML(html)
          links = result.css(".channel_info a.channel_name")
          break if links.blank?

          attrs << links.map { |link|
            {
              code: link.attr(:href)[/(ch[0-9]+)/],
              name: link.text.strip
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
