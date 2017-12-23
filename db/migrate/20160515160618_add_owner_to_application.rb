# frozen_string_literal: true

class AddOwnerToApplication < ActiveRecord::Migration[4.2]
  def change
    add_column :oauth_applications, :owner_id, :integer
    add_column :oauth_applications, :owner_type, :string
    add_index :oauth_applications, [:owner_id, :owner_type]
  end
end
