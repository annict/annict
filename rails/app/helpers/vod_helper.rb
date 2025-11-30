# typed: false
# frozen_string_literal: true

module VodHelper
  def vod_title_url(channel_id, code)
    case channel_id
    when Channel::AMAZON_VIDEO_ID
      "https://www.amazon.co.jp/gp/video/detail/#{code}"
    when Channel::BANDAI_CHANNEL_ID
      "https://www.b-ch.com/ttl/index.php?ttl_c=#{code}"
    when Channel::D_ANIME_STORE_ID
      "https://animestore.docomo.ne.jp/animestore/ci_pc?workId=#{code}"
    when Channel::D_ANIME_STORE_NICONICO_ID
      if code.start_with? "so"
        "https://www.nicovideo.jp/watch/#{code}"
      else
        "https://www.nicovideo.jp/series/#{code}"
      end
    when Channel::NICONICO_CHANNEL_ID
      "https://ch.nicovideo.jp/#{code}"
    when Channel::NETFLIX_ID
      "https://www.netflix.com/title/#{code}"
    when Channel::ABEMA_VIDEO_ID
      "https://abema.tv/video/title/#{code}"
    end
  end

  module_function :vod_title_url
end
