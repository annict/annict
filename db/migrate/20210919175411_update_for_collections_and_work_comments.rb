# frozen_string_literal: true

class UpdateForCollectionsAndWorkComments < ActiveRecord::Migration[6.1]
  def change
    add_column :collections, :collection_items_count, :integer, null: false, default: 0
    rename_column :collections, :title, :name
    remove_column :collections, :aasm_state
    remove_column :collections, :impressions_count
    change_column_null :collections, :description, false
    change_column_default :collections, :description, ""
    add_index :collections, %i[user_id name], unique: true

    remove_column :collection_items, :title
    remove_column :collection_items, :comment
    remove_column :collection_items, :aasm_state
    remove_column :collection_items, :reactions_count

    add_column :library_entries, :note, :text, null: false, default: ""
  end
end
