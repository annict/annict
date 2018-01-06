# frozen_string_literal: true

namespace :vod do
  task fetch: :environment do
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
        channel = Channel.published.with_video_service.find(data[:channel_id])
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
