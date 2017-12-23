class CreateTips < ActiveRecord::Migration[4.2]
  def change
    create_table :tips do |t|
      t.integer :target, null: false
      t.string :partial_name, null: false
      t.string :title, null: false
      t.string :icon_name, null: false
      t.timestamps null: false
    end

    add_index :tips, :partial_name, unique: true
  end
end
