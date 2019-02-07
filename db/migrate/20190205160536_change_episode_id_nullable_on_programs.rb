# frozen_string_literal: true

class ChangeEpisodeIdNullableOnPrograms < ActiveRecord::Migration[5.2]
  def change
    change_column_null :programs, :episode_id, true
  end
end
