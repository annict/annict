class CreateAppeals < ActiveRecord::Migration[4.2]
  def change
    create_table :appeals do |t|
      t.integer     :user_id,  null: false
      t.integer     :work_id
      t.integer     :category, null: false
      t.text        :body
      t.timestamps
    end

    add_foreign_key :appeals, :users
    add_foreign_key :appeals, :works
  end
end
