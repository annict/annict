class CreateTips < ActiveRecord::Migration
  def change
    create_table :tips do |t|
      t.string :title, null: false
      t.string :partial_name, null: false
      t.integer :target, null: false
      t.timestamps null: false
    end
  end
end
