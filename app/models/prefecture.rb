# == Schema Information
#
# Table name: prefectures
#
#  id         :integer          not null, primary key
#  name       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_prefectures_on_name  (name) UNIQUE
#

class Prefecture < ActiveRecord::Base
  validates :name, presence: true
end
