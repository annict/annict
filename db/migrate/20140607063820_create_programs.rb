class CreatePrograms < ActiveRecord::Migration[4.2]
  def change
    create_table :programs do |t|
      t.integer  :channel_id, null: false
      t.integer  :episode_id, null: false
      t.integer  :work_id,    null: false
      t.datetime :started_at, null: false
      t.timestamps
    end

    add_foreign_key :programs, :channels
    add_foreign_key :programs, :episodes
    add_foreign_key :programs, :works
  end
end
