# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: seasons
#
#  id          :bigint           not null, primary key
#  name        :string(510)      not null
#  sort_number :integer          not null
#  year        :integer          not null
#  created_at  :timestamptz
#  updated_at  :timestamptz
#
# Indexes
#
#  index_seasons_on_sort_number    (sort_number) UNIQUE
#  index_seasons_on_year           (year)
#  index_seasons_on_year_and_name  (year,name) UNIQUE
#

class SeasonModel < ApplicationRecord
  self.table_name = :seasons
end
