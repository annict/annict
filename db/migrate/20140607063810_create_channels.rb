class CreateChannels < ActiveRecord::Migration
  def change
    create_table :channels do |t|
      t.integer :channel_group_id, null: false
      t.integer :sc_chid,          null: false
      t.string  :name,             null: false
      t.timestamps

      t.foreign_key :channel_groups, dependent: :delete
    end

    add_index :channels, :sc_chid, unique: true
  end
end
