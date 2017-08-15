# frozen_string_literal: true

namespace :tmp do
  task update_channels: :environment do
    ActiveRecord::Base.transaction do
      [
        "Amazonビデオ",
        "Netflix (JP)"
      ].each do |name|
        puts name
        Channel.where(name: name).first_or_create!(streaming_service: true, channel_group_id: 14)
      end

      sc_chids = [234, 107, 165]
      Channel.where(sc_chid: sc_chids).update_all(streaming_service: true)

      Channel.find_by(sc_chid: 132).update(aasm_state: :hidden)
    end
  end

  task set_streaming_links: :environment do
    ActiveRecord::Base.transaction do
      [
        {
          channel_id: 107,
          scraping_class: BandaiChannel,
          url: "http://www.b-ch.com/ttl/index.php?ttl_c="
        },
        {
          channel_id: 165,
          scraping_class: NicoNicoCh,
          url: "http://ch.nicovideo.jp/"
        },
        {
          channel_id: 241,
          scraping_class: DAnimeStore,
          url: "https://anime.dmkt-sp.jp/animestore/ci_pc?workId="
        },
        {
          channel_id: 243,
          scraping_class: AmazonVideo,
          url: "https://www.amazon.co.jp/dp/"
        }
      ].each do |data|
        channel = Channel.published.with_streaming_service.find(data[:channel_id])
        results = data[:scraping_class].scrape
        results.each do |result|
          work = Work.published.find_by(title: result[:title])
          next if work.blank?
          puts "#{channel.name}: #{work.title}"
          channel.program_details.where(work: work).first_or_create! do |sl|
            sl.url = "#{data[:url]}#{result[:unique_id]}"
          end
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
    puts "--- BandaiChannel"
    path = "/ttl/jpchar_list.php"
    puts "Accessing to: #{path}"
    html = self.class.get(path)
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
    puts "--- NicoNicoCh"
    data = []

    1.step do |i|
      path = "/portal/anime/list?page=#{i}"
      puts "Accessing to: #{path}"
      html = self.class.get(path)
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
    puts "--- DAnimeStore"
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
          path = "/animestore/rest/WS000108?#{query}"
          puts "Accessing to: #{path}"
          response = self.class.get(path)
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
    puts "--- AmazonVideo"
    data = []

    1.step do |page|
      query = [
        "rh=n%3A2351649051%2Cn%3A%212351650051%2Cn%3A2478407051%2Cp_n_entity_type%3A4174099051%2Cp_n_ways_to_watch%3A3746328051",
        "page=#{page}"
      ].join("&")
      path = "/s/ref=sr_pg_2?#{query}"
      puts "Accessing to: #{path}"
      headers = { "User-Agent": "Mozilla/5.0" }
      response = self.class.get(path, headers: headers)
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
