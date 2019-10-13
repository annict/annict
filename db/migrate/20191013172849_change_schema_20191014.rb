# frozen_string_literal: true

class ChangeSchema20191014 < ActiveRecord::Migration[6.0]
  def change
    enable_extension "citext"

    change_column :users, :username, :citext
    change_column :users, :email, :citext
  end
end
