# frozen_string_literal: true

class CreateCollectionItemAgreements < ActiveRecord::Migration[5.0]
  def change
    create_table :collection_item_agreements do |t|
      t.integer :user_id, null: false
      t.integer :collection_item_id, null: false
      t.timestamps null: false
    end

    add_index :collection_item_agreements, :user_id
    add_index :collection_item_agreements, :collection_item_id
    add_index :collection_item_agreements, %i(user_id collection_item_id),
      unique: true, name: "index_cia_on_uid_and_ciid"

    add_foreign_key :collection_item_agreements, :users
    add_foreign_key :collection_item_agreements, :collection_items
  end
end
