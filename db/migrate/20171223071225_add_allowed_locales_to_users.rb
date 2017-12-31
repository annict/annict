# frozen_string_literal: true

class AddAllowedLocalesToUsers < ActiveRecord::Migration[5.1]
  def change
    add_column :users, :allowed_locales, :string, array: true
    add_index :users, :allowed_locales, using: "gin"
  end
end
