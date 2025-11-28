# frozen_string_literal: true

class RebuildItems < ActiveRecord::Migration[5.1]
  def change
    drop_table :draft_items, force: true
    drop_table :items, force: true

    create_table :items do |t|
      t.string :title, null: false
      t.string :detail_page_url, null: false
      t.string :asin, null: false
      t.string :ean
      t.integer :amount
      t.string :currency_code, null: false, default: ""
      t.integer :offer_amount
      t.string :offer_currency_code, null: false, default: ""
      t.datetime :release_on
      t.string :manufacturer, null: false, default: ""
      t.attachment :thumbnail
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :items, :asin, unique: true
  end
end
