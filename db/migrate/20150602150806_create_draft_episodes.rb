class CreateDraftEpisodes < ActiveRecord::Migration
  def change
    create_table :draft_episodes do |t|
      t.integer :episode_id, null: false
      t.integer :work_id, null: false
      t.string :number
      t.integer :sort_number, default: 0, null: false
      t.string :title
      t.timestamps null: false
    end

    add_index :draft_episodes, :episode_id
    add_index :draft_episodes, :work_id
    add_foreign_key :draft_episodes, :episodes
    add_foreign_key :draft_episodes, :works
  end
end
