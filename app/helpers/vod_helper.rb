# frozen_string_literal: true

module VodHelper
  def vod_title_url(channel_id, code)
    case channel_id
    when Channel::AMAZON_VIDEO_ID
      "https://www.amazon.co.jp/dp/#{code}"
    when Channel::BANDAI_CHANNEL_ID
      "http://www.b-ch.com/ttl/index.php?ttl_c=#{code}"
    when Channel::D_ANIME_STORE_ID
      "https://animestore.docomo.ne.jp/animestore/ci_pc?workId=#{code}"
    when Channel::NICONICO_CHANNEL_ID
      "http://ch.nicovideo.jp/#{code}"
    when Channel::NETFLIX_ID
      "https://www.netflix.com/title/#{code}"
    when Channel::ABEMA_VIDEO_ID
      "https://abema.tv/video/title/#{code}"
    end
  end
end
