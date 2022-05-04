# frozen_string_literal: true

module Annict
  module VodImporter
    class DAnimeStore < Annict::VodImporter::Base
      include HTTParty

      base_uri "https://animestore.docomo.ne.jp"

      def self.import
        new.import
      end

      def import
        puts "--- DAnimeStore"
        attrs = []

        1.step(10) do |collection_key|
          1.step(5) do |consonant_key|
            1.step do |start|
              query = [
                "workTypeList=anime",
                "start=#{((start * 100) - 100) + 1}",
                "length=100",
                "initialCollectionKey=#{collection_key}",
                "consonantKey=#{consonant_key}"
              ].join("&")
              path = "/animestore/rest/WS000108?#{query}"
              puts "Accessing to: #{path}"
              response = self.class.get(path)

              json = JSON.parse(response.body)
              works = json.dig("data", "workList")
              break if works.blank?

              attrs << works.map { |work|
                {
                  code: work["workId"],
                  name: work["workInfo"]["workTitle"].strip
                }
              }

              sleep rand(1..5)
            end
          end
        end

        channel = Channel.find(Channel::D_ANIME_STORE_ID)
        create_vod_title!(channel, attrs.flatten)
      end
    end
  end
end
