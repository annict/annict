# frozen_string_literal: true

class MultipleRecordsService
  def initialize(user)
    @user = user
  end

  def save!(episode_ids)
    CreateMultipleEpisodeRecordsJob.perform_later(@user.id, episode_ids)
  end
end
