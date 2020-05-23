# frozen_string_literal: true

class CreateMultipleEpisodeRecordsJob < ApplicationJob
  include V4::GraphqlRunnable

  queue_as :default

  def perform(user_id, episode_ids)
    user = User.only_kept.find(user_id)
    episodes = Episode.only_kept.where(id: episode_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      episodes.each do |episode|
        CreateEpisodeRecordRepository.new(graphql_client: graphql_client(viewer: user)).create(episode: episode,params: {})
      end
    end
  end
end
