# frozen_string_literal: true

class CreateUserPrograms < ActiveRecord::Migration[6.0]
  def change
    create_table :user_programs do |t|
      t.references :user, null: false, foreign_key: true
      t.references :work, null: false, foreign_key: true
      t.references :program, null: false, foreign_key: true
      t.timestamps
    end

    add_index :user_programs, %i(user_id work_id), unique: true
    add_index :user_programs, %i(user_id program_id), unique: true
  end
end
