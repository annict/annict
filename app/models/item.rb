# == Schema Information
#
# Table name: items
#
#  id         :integer          not null, primary key
#  work_id    :integer
#  name       :string(510)      not null
#  url        :string(510)      not null
#  image_uid  :string(510)      not null
#  main       :boolean          not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  items_work_id_idx  (work_id)
#

class Item < ActiveRecord::Base
  dragonfly_accessor :image

  belongs_to :work
end
