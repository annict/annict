# frozen_string_literal: true

class MultipleRecordsService
  attr_reader :user

  def initialize(user)
    @user = user
  end

  def save!(episode_ids)
    # 括弧とカンマ、数字だけかを確認する
    regex = /\A\[([0-9]+,*)+\]\z/
    return unless episode_ids =~ regex

    episode_ids = episode_ids.gsub(/\[|\]/, "").split(",")
    episodes = Episode.where(id: episode_ids).order(:sort_number)

    episodes.each do |episode|
      episode.checkins.create(user: user, work: episode.work)
    end
  end
end
