# frozen_string_literal: true

namespace :tmp do
  task delete_multiple_episode_records_from_db_activities: :environment do
    DbActivity.where(action: "multiple_episodes.create").delete_all
  end
end
