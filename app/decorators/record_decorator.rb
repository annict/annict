# frozen_string_literal: true

class RecordDecorator < ApplicationDecorator
  def records_show_title
    if episode.present?
      I18n.t("head.title.records.show_episode",
        profile_name: profile_name, username: username, work_title: work_title, episode_title_with_number: episode_title_with_number)
    else
      I18n.t("head.title.records.show_work", profile_name: profile_name, username: username, work_title: work_title)
    end
  end

  def records_show_description
    if episode.present?
      I18n.t("head.meta.description.records.show_episode",
        profile_name: profile_name, username: username, work_title: work_title, episode_title_with_number: episode_title_with_number)
    else
      I18n.t("head.meta.description.records.show_work",
        profile_name: profile_name, username: username, work_title: work_title)
    end
  end

  def facebook_share_body
    return comment if comment.present?

    if user.locale == "ja"
      "見ました。"
    else
      "Watched."
    end
  end

  def twitter_share_body
    user = record.user
    title = record.comment.present? ? work_title.truncate(30) : work_title
    comment = record.comment.present? ? "#{record.comment} / " : ""
    episode_number = episode.present? ? " #{episode.decorate.local_number} " : ""
    utm = {
      utm_source: "twitter",
      utm_medium: "record_share",
      utm_campaign: user.username
    }
    share_url = "#{detail_url}?#{utm.to_query}"
    share_hashtag = record.work.hashtag_with_hash

    base_body = if user.locale == "ja"
      "%s#{title}#{episode_number}を見ました #{share_url} #{share_hashtag}"
    else
      "%sWatched: #{title}#{episode_number} #{share_url} #{share_hashtag}"
    end

    body = base_body % comment
    body_without_url = body.sub(share_url, '')
    return body if body_without_url.length <= 130

    comment = comment.truncate(comment.length - (body_without_url.length - 130)) + " / "
    base_body % comment
  end

  private

  def profile_name
    user.profile.name
  end

  def username
    user.username
  end

  def work_title
    work.decorate.local_title
  end

  def episode_title_with_number
    episode.decorate.title_with_number
  end
end