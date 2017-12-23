class CreateTwitterBots < ActiveRecord::Migration[4.2]
  def change
    create_table :twitter_bots do |t|
      t.string :name, null: false
      t.timestamps
    end

    add_index :twitter_bots, :name, unique: true
  end
end
