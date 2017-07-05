# frozen_string_literal: true

namespace :data_care do
  task :merge_episode, %i(base_episode_id episode_id) => :environment do |_, args|
    base_episode_id, episode_id = args.values_at(:base_episode_id, :episode_id)
    merge_episode = Annict::DataCare::MergeEpisode.new(base_episode_id, episode_id)
    merge_episode.run!
  end

  task delete_abandoned_records: :environment do
    Activity.find_each do |a|
      if a.recipient.blank? || a.trackable.blank?
        puts "activity #{a.id} will be deleted"
        a.destroy
      end
    end
  end
end
