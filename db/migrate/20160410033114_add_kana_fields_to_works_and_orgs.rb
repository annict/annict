# frozen_string_literal: true

class AddKanaFieldsToWorksAndOrgs < ActiveRecord::Migration
  def change
    add_column :works, :title_kana, :string, null: false, default: ""
    add_column :organizations, :name_kana, :string, null: false, default: ""
    change_column_default :people, :name_kana, ""
    change_column_null :people, :name_kana, false, ""
  end
end
