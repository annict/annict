class CreateEpisodes < ActiveRecord::Migration
  def change
    create_table :episodes do |t|
      t.integer     :work_id,     null: false
      t.foreign_key :works,       dependent: :delete
      t.string      :number,      null: false
      t.integer     :sort_number, null: false, default: 0
      t.string      :title,       null: false, default: ''
      t.timestamps
    end
  end
end
