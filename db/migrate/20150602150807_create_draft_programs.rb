class CreateDraftPrograms < ActiveRecord::Migration
  def change
    create_table :draft_programs do |t|
      t.integer :program_id
      t.integer :channel_id, null: false
      t.integer :episode_id, null: false
      t.integer :work_id, null: false
      t.datetime :started_at, null: false
      t.timestamps null: false
    end

    add_index :draft_programs, :program_id
    add_index :draft_programs, :channel_id
    add_index :draft_programs, :episode_id
    add_index :draft_programs, :work_id
    add_foreign_key :draft_programs, :programs
    add_foreign_key :draft_programs, :channels
    add_foreign_key :draft_programs, :episodes
    add_foreign_key :draft_programs, :works
  end
end
