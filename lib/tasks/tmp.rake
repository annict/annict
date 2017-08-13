# frozen_string_literal: true

namespace :tmp do
  task move_from_collection_to_tag: :environment do
    ActiveRecord::Base.transaction do
      Collection.find_each do |c|
        puts "collection: #{c.id}"

        work_tag = WorkTag.where(name: c.title).first_or_create!
        c.user.work_taggables.where(work_tag: work_tag, description: c.description).first_or_create!

        c.collection_items.each do |ci|
          c.user.work_taggings.where(work: ci.work, work_tag: work_tag).first_or_create!
        end
      end
    end
  end

  task move_from_collection_item_to_work_comment: :environment do
    ActiveRecord::Base.transaction do
      CollectionItem.find_each do |ci|
        puts "collection item: #{ci.id}"

        work_comment = WorkComment.where(user: ci.user, work: ci.work).first_or_initialize

        if work_comment.body.present? && !work_comment.body.include?(ci.comment)
          work_comment.body += "\n#{ci.comment}"
        else
          work_comment.body = ci.comment
        end

        work_comment.save! if work_comment.body.present?
      end
    end
  end

  task update_channels: :environment do
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
