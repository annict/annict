# == Schema Information
#
# Table name: shots
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  image_uid  :string(255)      not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  index_shots_on_image_uid  (image_uid) UNIQUE
#

class Shot < ActiveRecord::Base
  dragonfly_accessor :image

  belongs_to :user
end
