# frozen_string_literal: true

class TwitterShareJob < ApplicationJob
  queue_as :default

  def perform(user_id, record_id)
    user = User.find(user_id)
    record = user.records.find(record_id)
    work = record.work
    episode = record.episode

    data = {
      work_title: work.decorate.local_title,
      episode_number: episode.single? ? "" : episode.decorate.local_number,
      share_hashtag: work.twitter_hashtag.blank? ? "" : " ##{work.twitter_hashtag}",
      comment: record.comment,
      locale: user.locale
    }

    TwitterService.new(user).share!(record, data)
  end
end
