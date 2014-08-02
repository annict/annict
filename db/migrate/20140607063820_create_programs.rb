class CreatePrograms < ActiveRecord::Migration
  def change
    create_table :programs do |t|
      t.integer  :channel_id, null: false
      t.integer  :episode_id, null: false
      t.integer  :work_id,    null: false
      t.datetime :started_at, null: false
      t.timestamps

      t.foreign_key :channels, dependent: :delete
      t.foreign_key :episodes, dependent: :delete
      t.foreign_key :works,    dependent: :delete
    end

    add_index :programs, [:channel_id, :episode_id], unique: true
  end
end
