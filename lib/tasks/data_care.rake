# frozen_string_literal: true

namespace :data_care do
  task :merge_episode, [:base_episode_id, :episode_id] => :environment do |_, args|
    base_episode_id, episode_id = args.values_at(:base_episode_id, :episode_id)
    merge_episode = Annict::DataCare::MergeEpisode.new(base_episode_id, episode_id)
    merge_episode.run!
  end
end
