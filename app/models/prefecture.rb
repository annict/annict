# typed: false

# == Schema Information
#
# Table name: prefectures
#
#  id         :bigint           not null, primary key
#  name       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_prefectures_on_name  (name) UNIQUE
#

class Prefecture < ApplicationRecord
  validates :name, presence: true
end
