# == Schema Information
#
# Table name: seasons
#
#  id         :integer          not null, primary key
#  name       :string(510)      not null
#  slug       :string(510)      not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  seasons_slug_key  (slug) UNIQUE
#

class Season < ActiveRecord::Base
  has_many :works
end
