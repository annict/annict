# frozen_string_literal: true

class V4 < ActiveRecord::Migration[6.0]
  def change
    %i(
      casts
      channel_groups
      channels
      characters
      episodes
      organizations
      people
      programs
      series
      series_works
      slots
      staffs
      trailers
      vod_titles
      works
    ).each do |table_name|
      add_column table_name, :unrevealed_at, :datetime
      add_index table_name, :unrevealed_at
    end
  end
end
