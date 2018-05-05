# frozen_string_literal: true

class StatusDecorator < ApplicationDecorator
  def facebook_share_body
    I18n.locale = user.locale
    work_title = work.decorate.local_title

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました。"
    else
      "Changed %s's status to \"#{kind_text}\"."
    end

    base_body % work_title
  end

  def twitter_share_body
    I18n.locale = user.locale
    work_title = work.decorate.local_title
    utm = {
      utm_source: "twitter",
      utm_medium: "status_share",
      utm_campaign: user.username
    }
    share_url = "#{library_url}?#{utm.to_query}"

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました。 #{share_url}"
    else
      "Changed %s's status to \"#{kind_text}\". Anime list: #{share_url}"
    end

    base_body % work_title
  end
end
