# frozen_string_literal: true

class ChangeDefaultScoreOnEpisodes < ActiveRecord::Migration[5.1]
  def change
    rename_column :episodes, :score, :satisfaction_score
    change_column_null :episodes, :satisfaction_score, true
    change_column_default :episodes, :satisfaction_score, nil

    rename_column :episodes, :avg_rating, :rating_avg

    add_column :works, :satisfaction_score, :float
    add_index :works, :satisfaction_score
  end
end
