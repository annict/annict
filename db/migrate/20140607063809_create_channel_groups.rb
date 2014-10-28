class CreateChannelGroups < ActiveRecord::Migration
  def change
    create_table :channel_groups do |t|
      t.string  :sc_chgid,    null: false
      t.string  :name,        null: false
      t.integer :sort_number
      t.timestamps
    end

    add_index :channel_groups, :sc_chgid, unique: true
  end
end
