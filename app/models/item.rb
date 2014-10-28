# == Schema Information
#
# Table name: items
#
#  id         :integer          not null, primary key
#  work_id    :integer
#  name       :string(255)      not null
#  url        :string(255)      not null
#  image_uid  :string(255)      not null
#  main       :boolean          default(FALSE), not null
#  created_at :datetime
#  updated_at :datetime
#

class Item < ActiveRecord::Base
  dragonfly_accessor :image

  belongs_to :work
end
