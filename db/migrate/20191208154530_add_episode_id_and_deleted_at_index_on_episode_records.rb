# frozen_string_literal: true

class AddEpisodeIdAndDeletedAtIndexOnEpisodeRecords < ActiveRecord::Migration[6.0]
  def change
    add_index :episode_records, %i[episode_id deleted_at]
    remove_index :episode_records, name: :checkins_episode_id_idx
    remove_index :episode_records, name: :index_episode_records_on_deleted_at
  end
end
