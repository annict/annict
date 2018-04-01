# frozen_string_literal: true

class StatusDecorator < ApplicationDecorator
  # Do not use helper methods via Draper when the method is used in ActiveJob
  # https://github.com/drapergem/draper/issues/655
  def library_url
    "#{user.annict_url}/@#{user.username}/#{kind}"
  end

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

  def tweet_body
    I18n.locale = user.locale
    work_title = work.decorate.local_title
    utm = {
      utm_source: "twitter",
      utm_medium: "status_share",
      utm_campaign: user.username
    }
    library_url_with_query = "#{library_url}?#{utm.to_query}"

    base_body = if user.locale == "ja"
      "アニメ「%s」の視聴ステータスを「#{kind_text}」にしました。 #{library_url_with_query}"
    else
      "Changed %s's status to \"#{kind_text}\". Anime list: #{library_url}"
    end

    base_body % work_title
  end
end
