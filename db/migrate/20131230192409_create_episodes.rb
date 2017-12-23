class CreateEpisodes < ActiveRecord::Migration[4.2]
  def change
    create_table :episodes do |t|
      t.integer     :work_id,     null: false
      t.string      :number,      null: false
      t.integer     :sort_number, null: false, default: 0
      t.string      :title,       null: false, default: ''
      t.timestamps
    end

    add_foreign_key :episodes, :works
  end
end
