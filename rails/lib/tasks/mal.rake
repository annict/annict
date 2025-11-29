# typed: false
# frozen_string_literal: true

namespace :mal do
  task import: :environment do
    mal = Mal.new
    mal.import!
  end

  class Mal
    include HTTParty

    base_uri "https://hbv3-mal-api.herokuapp.com/2.1"

    def import!
      page = 1

      loop do
        puts "Page: #{page}"
        anime_list = fetch_anime_list(page: page)
        break if anime_list.key?("error")

        anime_list.each do |anime|
          print "Anime ##{anime["id"]}: "

          if Work.find_by(mal_anime_id: anime["id"]).present?
            puts "already saved"
            next
          end

          mal_anime = fetch_anime(anime["id"])

          if mal_anime.key?("error")
            puts "error"
            next
          end

          work = find_work(mal_anime)

          if work.blank?
            puts "not found"
            next
          end

          import_work!(work, mal_anime)
        end

        page += 1
      end
    end

    def import_work!(work, mal_anime)
      attrs = {
        title_en: mal_anime["title"],
        mal_anime_id: mal_anime["id"]
      }
      work.update_columns(attrs)
      puts "updated"
    end

    private

    def fetch_anime_list(page: 1)
      self.class.get("/anime/popular", query: {page: page})
    end

    def fetch_anime(mal_anime_id)
      self.class.get("/anime/#{mal_anime_id}")
    end

    def find_work(mal_anime)
      work = Work.where("lower(title) = ?", mal_anime["title"].downcase).first

      if work.blank?
        (mal_anime.dig("other_titles", "japanese").presence || []).each do |title|
          work = Work.find_by(title: title)
        end
      end

      work
    end
  end
end
