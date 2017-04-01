# frozen_string_literal: true

namespace :tmp do
  task update_channels: :environment do
    [
      "Amazonビデオ",
      "Crunchyroll",
      "DAISUKI",
      "Funimation",
      "Hulu",
      "Netflix",
      "U-NEXT"
    ].each do |name|
      puts name
      Channel.where(name: name).first_or_create!(streaming_service: true)
    end

    sc_chids = [234, 107, 165]
    Channel.where(sc_chid: sc_chids).update_all(streaming_service: true)

    Channel.find_by(sc_chid: 132).update(aasm_state: :hidden)
  end

  task set_streaming_links: :environment do
    [
      {
        channel_id: 107,
        scraping_class: BandaiChannel
      },
      {
        channel_id: 165,
        scraping_class: NicoNicoCh
      },
      {
        channel_id: 241,
        scraping_class: DAnimeStore
      },
      {
        channel_id: 243,
        scraping_class: AmazonVideo
      },
      {
        channel_id: 247,
        scraping_class: Hulu
      },
      {
        channel_id: 249,
        scraping_class: UNext
      }
    ].each do |data|
      channel = Channel.published.with_streaming_service.find(data[:channel_id])
      results = data[:scraping_class].scrape
      results.each do |result|
        work = Work.published.find_by(title: result[:title])
        next if work.blank?
        puts "#{channel.name}: #{work.title}"
        channel.streaming_links.where(work: work, locale: "ja").first_or_create! do |sl|
          sl.unique_id = result[:unique_id]
        end
      end
    end
  end
end

class BandaiChannel
  include HTTParty

  base_uri "http://www.b-ch.com/"

  def self.scrape
    new.scrape
  end

  def scrape
    html = self.class.get("/ttl/jpchar_list.php")
    result = Nokogiri::HTML(html)
    links = result.css(".search-list .ttl-list li a")
    links.map do |link|
      {
        unique_id: link.attr(:href)[/([0-9]+)/],
        title: link.text
      }
    end
  end
end

class NicoNicoCh
  include HTTParty

  base_uri "http://ch.nicovideo.jp/"

  def self.scrape
    new.scrape
  end

  def scrape
    data = []

    1.step do |i|
      html = self.class.get("/portal/anime/list?page=#{i}")
      result = Nokogiri::HTML(html)
      links = result.css(".channel_info a.channel_name")
      break if links.blank?
      data << links.map do |link|
        {
          unique_id: link.attr(:href)[/(ch[0-9]+)/],
          title: link.text.strip
        }
      end
      sleep rand(1..5)
    end

    data.flatten
  end
end

class DAnimeStore
  include HTTParty

  base_uri "https://anime.dmkt-sp.jp"

  def self.scrape
    new.scrape
  end

  def scrape
    data = []

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
          puts "query: #{query}"
          response = self.class.get("/animestore/rest/WS000108?#{query}")
          json = JSON.parse(response.body)
          works = json.dig("data", "workList")
          break if works.blank?
          data << works.map do |work|
            {
              unique_id: work["workId"],
              title: work["workInfo"]["workTitle"]
            }
          end
          sleep rand(1..5)
        end
      end
    end

    data.flatten
  end
end

class AmazonVideo
  include HTTParty

  base_uri "https://www.amazon.co.jp"

  def self.scrape
    new.scrape
  end

  def scrape
    data = []

    1.step do |page|
      query = [
        "rh=n%3A2351649051%2Cn%3A%212351650051%2Cn%3A2478407051%2Cp_n_entity_type%3A4174099051%2Cp_n_ways_to_watch%3A3746328051",
        "page=#{page}"
      ].join("&")
      puts "page: #{page}"
      headers = { "User-Agent": "Mozilla/5.0" }
      response = self.class.get("/s/ref=sr_pg_2?#{query}", headers: headers)
      result = Nokogiri::HTML(response)
      item_list = result.css(".s-result-item")
      break if item_list.blank?
      data << item_list.map do |item|
        {
          unique_id: item.attr("data-asin"),
          title: item.css(".s-access-title").text
        }
      end
      sleep rand(1..5)
    end

    data.flatten
  end
end

class Hulu
  include HTTParty

  base_uri "http://www.hulu.jp"

  def self.scrape
    new.scrape
  end

  def scrape
    data = []

    1.step do |page|
      query = [
        "asset_scope=&featureable_scope=null&_language=ja&_region=jp",
        "items_per_page=100",
        "position=#{((page * 100) - 100) + 1}",
        "_user_pgid=8&_content_pgid=24&_device_id=1&region=jp&locale=ja&language=ja",
        "access_token=ysNpI_Xa59mzo99e7x_TROQVmhY-2m55M7O8GtH241kXPfU0_g--CY223ZHI7LH3lXDV5wnhk9F6mIbFOpn5ttseaYDqJsabRCLM6KZVmO98n_03qvxJjc193wyL9WjKUmu26Nd_DsNkKSzhzwOmr/2nuFDjAYoat1AmDgHGIQvo8q5TYerBrXxgNQufarOw7S1vxplvEB9SouhSgACsh0C5Qn9LuuEWINen8Nvqyk0Da4BnNShJx5RD20Oor9fKUTVoGzLpmGM8L3T6rqmBK4ITihemSxlurKGxGffxkCmQaVTe/9oywGPznKkbqumqZ1g_te1Ljg--"
      ].join("&")
      puts "page: #{page}"
      response = self.class.get("/mozart/v1.h2o/editorial/3963?#{query}")
      json = JSON.parse(response.body)
      works = json.dig("data")
      break if works.blank?
      data << works.map do |work|
        {
          unique_id: work["show"]["canonical_name"],
          title: work["show"]["name"]
        }
      end
      sleep rand(1..5)
    end

    data.flatten
  end
end

class UNext
  include HTTParty

  base_uri "http://video.unext.jp"

  def self.scrape
    new.scrape
  end

  def scrape
    data = []
    first_work = {}

    1.step do |page|
      query = [
        "filter=&order=popular&genre=MNU0000136&category=MNU0000768&page=#{page}"
      ].join("&")
      puts "page: #{page}"
      response = self.class.get("/api/1/search/video?#{query}")
      json = JSON.parse(response.body)
      works = json.dig("data", "title_list")
      break if first_work["title_code"] == works.first["title_code"]
      first_work = works.first
      data << works.map do |work|
        {
          unique_id: work["title_code"],
          title: work["title_name"]
        }
      end
      sleep rand(1..5)
    end

    data.flatten
  end
end
