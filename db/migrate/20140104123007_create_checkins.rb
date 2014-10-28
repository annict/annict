class CreateCheckins < ActiveRecord::Migration
  def change
    create_table :checkins do |t|
      t.integer     :user_id,             null: false
      t.foreign_key :users,               dependent: :delete
      t.integer     :episode_id,          null: false
      t.foreign_key :episodes,            dependent: :delete
      t.text        :comment
      t.string      :twitter_url_hash,    null: false, default: ''
      t.integer     :twitter_click_count, null: false, default: 0
      t.timestamps
    end
  end
end
