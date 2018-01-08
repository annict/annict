# frozen_string_literal: true

namespace :tmp do
  task set_vod_title_code: :environment do
    ProgramDetail.where.not(url: nil).find_each do |pd|
      print "ProductDetail: #{pd.id} -> "

      code = case pd.url
      when /netflix\.com/
        pd.url.scan(%r{https://www.netflix.com/title/([0-9]+)})&.last&.first
      when /b-ch\.com/
        pd.url.scan(%r{http://www.b-ch.com/ttl/index.php\?ttl_c=([0-9]+)})&.last&.first.presence ||
          pd.url.scan(%r{http://www.b-ch.com/ttl/index_html5.php\?ttl_c=([0-9]+)})&.last&.first.presence
      when /ch\.nicovideo\.jp/
        pd.url.scan(%r{http://ch.nicovideo.jp/(ch[0-9]+)})&.last&.first.presence ||
          pd.url.scan(%r{http://ch.nicovideo.jp/([0-9a-zA-Z_-]+)})&.last&.first
      when /anime\.dmkt-sp\.jp/
        pd.url.scan(%r{https://anime.dmkt-sp.jp/animestore/ci_pc\?workId=([0-9]+)})&.last&.first
      when /www\.amazon\.co\.jp/
        pd.url.scan(%r{https://www.amazon.co.jp/dp/([0-9a-zA-Z]+)})&.last&.first.presence ||
          pd.url.scan(%r{https://www.amazon.co.jp/gp/video/detail/([0-9a-zA-Z]+)})&.last&.first
      end

      if code.nil?
        puts "nil"
        next
      end

      puts "code: #{code}"
      pd.update_columns(url: nil, vod_title_code: code, vod_title_name: pd.work.title)
    end
  end
end
