class CreateAppeals < ActiveRecord::Migration
  def change
    create_table :appeals do |t|
      t.integer     :user_id,  null: false
      t.foreign_key :users,    dependent: :delete
      t.integer     :work_id
      t.foreign_key :works,    dependent: :delete
      t.integer     :category, null: false
      t.text        :body
      t.timestamps
    end
  end
end
