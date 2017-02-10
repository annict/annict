# frozen_string_literal: true

class RemoveIndexCastsOnWorkIdAndCharacterId < ActiveRecord::Migration[5.0]
  def change
    remove_index :casts, name: :index_casts_on_work_id_and_character_id
  end
end
