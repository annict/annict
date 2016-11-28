# frozen_string_literal: true

class CreateDislikes < ActiveRecord::Migration[5.0]
  def change
    create_table :dislikes do |t|
      t.integer :user_id, null: false
      t.references :recipient, null: false, polymorphic: true
      t.timestamps null: false
    end

    add_index :dislikes, :user_id
    add_index :dislikes, %i(recipient_id recipient_type)

    add_foreign_key :dislikes, :users
  end
end
