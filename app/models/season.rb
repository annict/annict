# == Schema Information
#
# Table name: seasons
#
#  id          :integer          not null, primary key
#  name        :string(510)      not null
#  slug        :string(510)      not null
#  created_at  :datetime
#  updated_at  :datetime
#  sort_number :integer
#
# Indexes
#
#  index_seasons_on_sort_number  (sort_number) UNIQUE
#  seasons_slug_key              (slug) UNIQUE
#

class Season < ActiveRecord::Base
  has_many :works
end
