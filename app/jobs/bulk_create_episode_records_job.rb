# frozen_string_literal: true

class BulkCreateEpisodeRecordsJob < ApplicationJob
  include V4::GraphqlRunnable

  queue_as :default

  def perform(user_id, episode_ids)
    user = User.only_kept.find(user_id)
    episode_database_ids = episode_ids.map { |episode_id| GraphQL::Schema::UniqueWithinType.decode(episode_id)[1] }.compact
    episodes = Episode.only_kept.where(id: episode_database_ids).order(:sort_number)

    return if episodes.blank?

    ActiveRecord::Base.transaction do
      episodes.each do |episode|
        episode_id = Canary::AnnictSchema.id_from_object(episode, episode.class)
        form = EpisodeRecordForm.new(episode_id: episode_id)

        episode_record, err = CreateEpisodeRecordRepository.new(
          graphql_client: graphql_client(viewer: user)
        ).execute(form: form)
      end
    end
  end
end
