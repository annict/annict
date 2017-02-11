# frozen_string_literal: true

class FixConstraintsOnSomeTables < ActiveRecord::Migration[5.0]
  def change
    add_foreign_key :db_comments, :users
    change_column_default :users, :locale, nil
    change_column_default :users, :time_zone, nil
    change_column_null :casts, :character_id, false
  end
end
