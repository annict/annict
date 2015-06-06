class CreateDraftMultipleEpisodes < ActiveRecord::Migration
  def change
    create_table :draft_multiple_episodes do |t|
      t.integer :work_id, null: false
      t.text :body, null: false
      t.timestamps null: false
    end

    add_index :draft_multiple_episodes, :work_id
    add_foreign_key :draft_multiple_episodes, :works
  end
end
