class CreateSyobocalAlerts < ActiveRecord::Migration
  def change
    create_table :syobocal_alerts do |t|
      t.integer :work_id
      t.integer :kind, null: false
      t.integer :sc_prog_item_id
      t.string :sc_sub_title
      t.string :sc_prog_comment
      t.timestamps

      t.foreign_key :works, dependent: :delete
    end

    add_index :syobocal_alerts, :kind
    add_index :syobocal_alerts, :sc_prog_item_id
  end
end
