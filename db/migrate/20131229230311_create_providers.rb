class CreateProviders < ActiveRecord::Migration
  def change
    create_table :providers do |t|
      t.integer     :user_id,         null: false
      t.foreign_key :users,           dependent: :delete
      t.string      :name,            null: false
      t.string      :uid,             null: false
      t.string      :token,           null: false
      t.integer     :token_expires_at
      t.string      :token_secret,    null: ''
      t.timestamps
    end
  end
end
