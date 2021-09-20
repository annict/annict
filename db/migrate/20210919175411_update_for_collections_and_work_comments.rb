# frozen_string_literal: true

class UpdateForCollectionsAndWorkComments < ActiveRecord::Migration[6.1]
  def change
    add_column :collections, :collection_items_count, :integer, null: false, default: 0
    remove_column :collections, :aasm_state
    remove_column :collections, :impressions_count
    change_column_null :collections, :description, false
    change_column_default :collections, :description, ""

    remove_column :collection_items, :aasm_state
    remove_column :collection_items, :title
    rename_column :collection_items, :comment, :body
    rename_column :collection_items, :reactions_count, :likes_count
    change_column_null :collection_items, :body, false
    change_column_default :collection_items, :body, ""

    add_column :library_entries, :note, :text, null: false, default: ""
  end
end
